package authsession

import (
	"context"
	"encoding/json"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/transport"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	DefaultPingTimeout  = 60
	PingAddTimeout      = 15
	CacheSessionTimeout = 3 * 60
	WaitMsgAckTimeout   = 30
)

const (
	StateNew = iota
	StateOnline
	StateOffline
	StateClose
)

const (
	SessionStateNew = iota
	SessionStateCreated
)

const (
	MsgIdTooLow  = int32(16)
	MsgIdTooHigh = int32(17)
)

type messageData struct {
	confirmFlag  bool
	compressFlag bool
	clientMsgId  int64
	obj          []byte
}

func (m *messageData) String() string {
	return fmt.Sprintf("{confirmFlag: %v, compressFlag: %v, obj: {%s}}", m.confirmFlag, m.compressFlag, m.obj)
}

type serverIdCtx struct {
	gatewayId       string
	lastReceiveTime int64
}

func (c serverIdCtx) Equal(id string) bool {
	return c.gatewayId == id
}

type session struct {
	sessionId       int64
	sessionState    int
	gatewayId       *serverIdCtx
	nextSeqNo       uint32
	firstMsgId      int64
	connState       int
	closeDate       int64
	lastReceiveTime int64
	inQueue         *sessionInboundQueue
	outQueue        *sessionOutgoingQueue
	sessList        *SessionList
}

func newSession(sessionId int64, sessList *SessionList) *session {
	sess := &session{
		sessionId:       sessionId,
		gatewayId:       nil,
		sessionState:    SessionStateNew,
		closeDate:       time.Now().Unix() + DefaultPingTimeout + PingAddTimeout,
		connState:       StateNew,
		lastReceiveTime: time.Now().UnixNano(),
		inQueue:         newSessionInboundQueue(),
		outQueue:        newSessionOutgoingQueue(),
		sessList:        sessList,
	}

	return sess
}

func (c *session) String() string {
	return fmt.Sprintf("{user_id: %d, auth_key_id: %d, session_id: %d, state: %d, conn_state: %d, conn_id_list: %#v}",
		c.sessList.cb.AuthUserId,
		c.sessList.authId,
		c.sessionId,
		c.sessionState,
		c.connState,
		c.gatewayId)
}

func (c *session) setGatewayId(gateId string) {
	if c.gatewayId == nil {
		c.gatewayId = &serverIdCtx{gatewayId: gateId, lastReceiveTime: time.Now().Unix()}
	} else {
		c.gatewayId.gatewayId = gateId
		c.gatewayId.lastReceiveTime = time.Now().Unix()
	}
}

func (c *session) getGatewayId() string {
	if c.gatewayId == nil {
		return ""
	} else {
		return c.gatewayId.gatewayId
	}
}

func (c *session) checkGatewayIdExist(gateId string) bool {
	if c.gatewayId == nil {
		return false
	}

	return c.gatewayId.Equal(gateId)
}

func (c *session) changeConnState(ctx context.Context, state int) {
	c.connState = state
	if state == StateOnline {
		c.sessList.cb.setOnline(ctx)
	} else if state == StateOffline {
		c.sessList.cb.trySetOffline(ctx)
	}
}

func (c *session) onSessionConnNew(ctx context.Context, gatewayId string) {
	if c.connState != StateOnline {
		c.changeConnState(ctx, StateOnline)
		c.setGatewayId(gatewayId)
	}

	// TODO: fake onSetMainUpdatesSession
	c.sessList.cb.onSetMainUpdatesSession(ctx, c)
	c.sessList.cb.changeAuthState(ctx, AuthStateNormal, int64(1))
}

func (c *session) onSessionMessageData(ctx context.Context, gatewayId string, clientIp string, msg *transport.TMsgRawData) {
	willCloseDate := time.Now().Unix() + DefaultPingTimeout + PingAddTimeout
	if willCloseDate > c.closeDate {
		c.closeDate = willCloseDate
	}

	if !c.checkBadMsgNotification(ctx, gatewayId, false, msg) {
		logx.WithContext(ctx).Errorf("badMsgNotification - {sess: %s, conn_id: %s}", c, gatewayId)
		return
	}

	var (
		messages []*transport.TMsgRawData
	)

	messages = append(messages, msg)

	// check onNewSessionCreated
	minMsgId := msg.MsgId
	for _, m2 := range messages {
		if minMsgId < m2.MsgId {
			minMsgId = m2.MsgId
		}
	}

	if c.sessionState == SessionStateNew || minMsgId < c.firstMsgId {
		logx.WithContext(ctx).Infof("onNewSessionCreated - %#v, c: %s", messages, c)
		if c.firstMsgId != 0 {
			c.firstMsgId = minMsgId
		}
		c.sessionState = SessionStateCreated
	}

	defer func() {
		c.sendQueueToClient(ctx, gatewayId)
		c.inQueue.Shrink()
	}()

	for _, m2 := range messages {
		if m2.MsgId < c.firstMsgId {
			continue
		}
		if !c.checkBadMsgNotification(ctx, gatewayId, true, m2) {
			continue
		}

		if m2.Body == nil && m2.Event == "" {
			logx.WithContext(ctx).Errorf("obj is nil: %v", m2)
			continue
		}

		if m2.Event == "ack" {
			// unmarshal and ack
			c.onMsgAck(ctx, gatewayId, m2.MsgId, []int64{})
		} else {
			inMsg := c.inQueue.AddMsgId(m2.MsgId)
			if inMsg.state == NONE {
				c.onEventRequest(ctx, gatewayId, clientIp, inMsg, m2.Event, m2.Body)
			} else {
				continue
			}
		}
	}
}

func (c *session) onSessionConnClose(ctx context.Context, id string) {
	if c.checkGatewayIdExist(id) {
		c.gatewayId = nil
		c.changeConnState(ctx, StateOffline)
	}
}

func (c *session) sessionOnline() bool {
	return c.connState == StateOnline
}

func (c *session) sessionClosed() bool {
	return c.connState == StateClose
}

func (c *session) onTimer(ctx context.Context) bool {
	date := time.Now().Unix()
	// log.Debugf("onTimer - c: %s, outQ len: %d", c.String(), c.outQueue.listOutMsg.Len())
	gatewayId := c.getGatewayId()

	if c.connState == StateOnline {
		if date >= c.closeDate {
			c.changeConnState(context.Background(), StateOffline)
		} else {
			c.sendQueueToClient(ctx, gatewayId)
		}
	} else if c.connState == StateOffline || c.connState == StateNew {
		if date >= c.closeDate+CacheSessionTimeout {
			c.changeConnState(context.Background(), StateClose)
		}
	}
	return true
}

func (c *session) generateMessageSeqNo(increment bool) int32 {
	value := c.nextSeqNo
	logx.Info("generateMessageSeqNo - value: ", value)
	if increment {
		c.nextSeqNo++
		return int32(value*2 + 1)
	} else {
		return int32(value * 2)
	}
}

func (c *session) sendPushToQueue(ctx context.Context, gatewayId string, pushMsgId int64, reqMsgId int64, event string, pushMsg []byte) {

	//if x.GetOffset() > 256 {
	//	gzipPacked := &tproto.TGzipPacked{
	//		PackedData: rawBytes,
	//	}
	//	x2 := tproto.NewEncodeBuf(512)
	//	_ = gzipPacked.Encode(x2)
	//	rawBytes = x2.GetBuf()
	//}

	rawMsg := &transport.TMsgRawData{
		MsgId: nextMessageId(),
		//SeqNo:    c.generateMessageSeqNo(true),
		ReqMsgId: reqMsgId,
		Event:    event,
		Body:     pushMsg,
	}
	c.outQueue.AddPushUpdates(pushMsgId, rawMsg)
}

func (c *session) sendRawToQueue(ctx context.Context, gatewayId string, reqMsgId int64, confirm bool, event string, raw []byte) {

	rawMsg2 := &transport.TMsgRawData{
		MsgId: nextMessageId(),
		//SeqNo:    c.generateMessageSeqNo(confirm),
		ReqMsgId: reqMsgId,
		Event:    event,
		Body:     raw,
	}
	c.outQueue.AddNotifyMessage(reqMsgId, confirm, rawMsg2)
}

func (c *session) sendRawDirectToGateway(ctx context.Context, gatewayId string, raw *transport.TMsgRawData) (bool, error) {
	logx.Debugf("sendRawDirectToGateway - gatewayId: %s, status: %v", gatewayId, c.connState)
	if c.connState != StateOnline {
		return false, nil
	}

	return c.sessList.cb.sendCb(
		ctx,
		gatewayId,
		c.sessList.authId,
		c.sessionId,
		raw)
}

func (c *session) sendDirectToGateway(ctx context.Context, gatewayId string, confirm bool, obj []byte) (bool, error) {
	if c.connState != StateOnline {
		return false, nil
	}

	rawMsg := &transport.TMsgRawData{
		MsgId: nextMessageId(),
		//SeqNo: c.generateMessageSeqNo(confirm),
		Body: obj,
	}

	return c.sessList.cb.sendCb(
		ctx,
		gatewayId,
		c.sessList.authId,
		c.sessionId,
		rawMsg)
}

func (c *session) sendQueueToClient(ctx context.Context, gatewayId string) {
	if gatewayId == "" {
		return
	}

	if c.outQueue.listOutMsg.Len() == 0 {
		return
	}

	var (
		listPending = make([]*outboxMsg, 0)
		b           bool
		err         error
		sentTime    = time.Now().Unix()
	)
	for e := c.outQueue.listOutMsg.Front(); e != nil; e = e.Next() {
		if e.Value.(*outboxMsg).sent == 0 || time.Now().Unix() >= e.Value.(*outboxMsg).sent+WaitMsgAckTimeout {
			listPending = append(listPending, e.Value.(*outboxMsg))
		}
	}

	for _, m := range listPending {
		logx.Debugf("ListPending: %+v - %d ", m.msgId, m.msg.MsgId)
	}

	if len(listPending) == 1 {
		logx.WithContext(ctx).Infof("sendRawDirectToGateway - pending[0]")
		b, err = c.sendRawDirectToGateway(ctx, gatewayId, listPending[0].msg)
		// log.Debugf("err: %v, b: %v", err, b)
		if err != nil || !b {
			return
		}

		for _, m := range listPending {
			if m.state == NO_NEED_ACK {
				logx.WithContext(ctx).Infof("no_need_ack msg_id: %d", m.msgId)
				c.outQueue.Remove(m.msgId)
			} else {
				logx.WithContext(ctx).Infof("pending sent: %d", m.msgId)
				m.sent = sentTime
			}
		}
	} else if len(listPending) > 1 {
		var (
			split = 8
		)
		for i := 0; i < len(listPending)/split; i++ {
			msgContainer := &transport.TMsgRawDataContainer{
				Messages: make([]*transport.TMsgRawData, 0, split),
			}
			for _, m := range listPending[i*split : (i+1)*split] {
				msgContainer.Messages = append(msgContainer.Messages, m.msg)
			}
			logx.WithContext(ctx).Infof("sendRawDirectToGateway - %+v", msgContainer)
			msgContainerData, _ := json.Marshal(msgContainer)
			b, err = c.sendDirectToGateway(ctx, gatewayId, false, msgContainerData)
			if err != nil || !b {
				continue
			}

			for _, m := range listPending[i*split : (i+1)*split] {
				if m.state == NO_NEED_ACK {
					logx.WithContext(ctx).Infof("need_no_ack: %d", m.msgId)
					c.outQueue.Remove(m.msgId)
				} else {
					logx.WithContext(ctx).Infof("pending sent: %d", m.msgId)
					m.sent = sentTime
				}
			}
		}
		if (len(listPending) % split) > 0 {
			msgContainer := &transport.TMsgRawDataContainer{
				Messages: make([]*transport.TMsgRawData, 0, split),
			}
			for _, m := range listPending[split*(len(listPending)/split):] {
				msgContainer.Messages = append(msgContainer.Messages, m.msg)
			}
			logx.WithContext(ctx).Infof("sendRawDirectToGateway - %+v", msgContainer)
			msgContainerData, _ := json.Marshal(msgContainer)
			b, err = c.sendDirectToGateway(ctx, gatewayId, false, msgContainerData)
			// log.Debugf("err: %v, b: %v", err, b)
			if err != nil || !b {
				return
			}
			for _, m := range listPending[split*(len(listPending)/split):] {
				if m.state == NO_NEED_ACK {
					logx.WithContext(ctx).Infof("need_no_ack: %d", m.msgId)
					c.outQueue.Remove(m.msgId)
				} else {
					logx.WithContext(ctx).Infof("pending sent: %d", m.msgId)
					m.sent = sentTime
				}
			}
		}
	}
}
