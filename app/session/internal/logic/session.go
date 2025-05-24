package logic

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
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
	obj          tproto.TObject
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
	pendingQueue    *sessionRpcResultWaitingQueue
	sessList        *SessionList
	isHttp          bool
	httpQueue       *httpRequestQueue
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
		pendingQueue:    newSessionRpcResultWaitingQueue(),
		sessList:        sessList,
		isHttp:          false,
		httpQueue:       newHttpRequestQueue(),
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

func (c *session) onSessionConnNew(ctx context.Context, id string) {
	if c.connState != StateOnline {
		c.changeConnState(ctx, StateOnline)
		c.setGatewayId(id)
	}
}

func (c *session) onSessionHttpMessageData(ctx context.Context, gatewayId, clientIp string, msg *tproto.TMessage) {

	if !c.checkBadMsgNotification(ctx, gatewayId, false, msg) {
		logx.WithContext(ctx).Errorf("badMsgNotification - {sess: %s, conn_id: %s}", c, gatewayId)
		return
	}

	var (
		messages []*tproto.TMessage
	)
	msgContainer := &tproto.TMsgContainer{}

	if err := msg.Object.UnmarshalTo(msgContainer); err == nil {
		messages = msgContainer.Messages

		c.inQueue.AddMsgId(msg.MsgId)
	} else {
		messages = append(messages, msg)
	}

	for i := 0; i < len(messages); i++ {
		gzipPacked := &tproto.TGzipPacked{}
		if err := messages[i].Object.UnmarshalTo(gzipPacked); err == nil {
			messages[i] = &tproto.TMessage{
				MsgId:  messages[i].MsgId,
				SeqNo:  messages[i].SeqNo,
				Bytes:  int32(len(gzipPacked.PackedData)),
				Object: gzipPacked.Obj,
			}
		}
	}

	// check onNewSessionCreated
	minMsgId := msg.MsgId
	for _, m2 := range messages {
		if minMsgId < m2.MsgId {
			minMsgId = m2.MsgId
		}
	}

	if c.sessionState == SessionStateNew || minMsgId < c.firstMsgId {
		logx.WithContext(ctx).Infof("onNewSessionCreated - %#v, c: %s", messages, c)
		c.onNewSessionCreated(ctx, gatewayId, minMsgId)
		if c.firstMsgId != 0 {
			c.firstMsgId = minMsgId
		}
		c.sessionState = SessionStateCreated
	}

	defer func() {
		c.sendQueueToGateway(ctx, gatewayId)
		c.inQueue.Shrink()
	}()

	for _, m2 := range messages {
		if m2.MsgId < c.firstMsgId {
			continue
		}
		if !c.checkBadMsgNotification(ctx, gatewayId, true, m2) {
			continue
		}

		if m2.Object == nil {
			logx.WithContext(ctx).Errorf("obj is nil: %v", m2)
			continue
		}

		inMsg := c.inQueue.AddMsgId(m2.MsgId)
		if inMsg.state == NONE {
			c.processMsg(ctx, gatewayId, clientIp, inMsg, m2.Object)
		} else {
			continue
		}
	}
}

func (c *session) onSessionMessageData(ctx context.Context, gatewayId, clientIp string, msg *tproto.TMessage) {
	willCloseDate := time.Now().Unix() + DefaultPingTimeout + PingAddTimeout
	if willCloseDate > c.closeDate {
		c.closeDate = willCloseDate
	}

	if !c.checkBadMsgNotification(ctx, gatewayId, false, msg) {
		logx.WithContext(ctx).Errorf("badMsgNotification - {sess: %s, conn_id: %s}", c, gatewayId)
		return
	}

	var (
		messages []*tproto.TMessage
	)
	msgContainer := &tproto.TMsgContainer{}

	if err := msg.Object.UnmarshalTo(msgContainer); err == nil {
		messages = msgContainer.Messages

		// check
		c.inQueue.AddMsgId(msg.MsgId)
	} else {
		messages = append(messages, msg)
	}

	for i := 0; i < len(messages); i++ {
		gzipPacked := &tproto.TGzipPacked{}
		if err := messages[i].Object.UnmarshalTo(gzipPacked); err == nil {
			messages[i] = &tproto.TMessage{
				MsgId:  messages[i].MsgId,
				SeqNo:  messages[i].SeqNo,
				Bytes:  int32(len(gzipPacked.PackedData)),
				Object: gzipPacked.Obj,
			}
		}
	}

	// check onNewSessionCreated
	minMsgId := msg.MsgId
	for _, m2 := range messages {
		if minMsgId < m2.MsgId {
			minMsgId = m2.MsgId
		}
	}

	if c.sessionState == SessionStateNew || minMsgId < c.firstMsgId {
		logx.WithContext(ctx).Infof("onNewSessionCreated - %#v, c: %s", messages, c)
		c.onNewSessionCreated(ctx, gatewayId, minMsgId)
		if c.firstMsgId != 0 {
			c.firstMsgId = minMsgId
		}
		c.sessionState = SessionStateCreated
	}

	defer func() {
		c.sessList.cb.sendToRpcQueue(ctx, c.sessList.cb.tmpRpcApiMessageList)
		c.sessList.cb.tmpRpcApiMessageList = []*rpcApiMessage{}

		c.sendQueueToGateway(ctx, gatewayId)
		c.inQueue.Shrink()
	}()

	for _, m2 := range messages {
		if m2.MsgId < c.firstMsgId {
			continue
		}
		if !c.checkBadMsgNotification(ctx, gatewayId, true, m2) {
			continue
		}

		if m2.Object == nil {
			logx.WithContext(ctx).Errorf("obj is nil: %v", m2)
			continue
		}

		if tproto.IsAnyOfType(m2.Object, &tproto.TMsgAck{}) {
			m2Ack := &tproto.TMsgAck{}
			if err := m2.Object.UnmarshalTo(m2Ack); err != nil {
				continue
			}
			c.onMsgAck(ctx, gatewayId, m2.MsgId, m2.SeqNo, m2Ack)
		} else {
			inMsg := c.inQueue.AddMsgId(m2.MsgId)
			if inMsg.state == NONE {
				c.processMsg(ctx, gatewayId, clientIp, inMsg, m2.Object)
			} else {
				continue
			}
		}
	}
}

func (c *session) processMsg(ctx context.Context, gatewayId, clientIp string, inMsg *inboxMsg, r *anypb.Any) {
	actualObj, _ := r.UnmarshalNew()
	switch a := actualObj.(type) {
	case *sessionpb.TSession_TTTTT_TestUser:
		c.onTestSetUser(ctx, gatewayId, inMsg, a)
	case *tproto.Ping:
		c.onPing(ctx, gatewayId, inMsg, a)
	case *tproto.InitConnection:
		c.onInitConnection(ctx, gatewayId, clientIp, inMsg, a)
	case *tproto.DestroySession:
		c.onDestroySession(ctx, gatewayId, inMsg, a)
	default:
		logx.Infof("session: %s, gatewayId: %s, clientIp: %s, inMsg: %s, a: %s", c, gatewayId, clientIp, inMsg, a)
		c.onRpcRequest(ctx, gatewayId, clientIp, inMsg, a)
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

	result, err := anypb.New(&tproto.RpcError{
		ErrorCode:    -503,
		ErrorMessage: "Timeout",
	})

	if err != nil {
		logx.WithContext(ctx).Errorf("onTimer - anypb.New error: %s", err.Error())
		return false
	}

	timeoutIdList := c.pendingQueue.OnTimer()
	for _, id := range timeoutIdList {
		c.sendRpcResult(
			ctx,
			&tproto.RpcResult{
				ReqMsgId: id,
				Result:   result,
			})
	}

	if c.connState == StateOnline {
		if date >= c.closeDate {
			c.changeConnState(context.Background(), StateOffline)
		} else {
			c.sendQueueToGateway(ctx, gatewayId)
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
	if increment {
		c.nextSeqNo++
		return int32(value*2 + 1)
	} else {
		return int32(value * 2)
	}
}

func (c *session) sendRpcResultToQueue(ctx context.Context, gatewayId string, reqMsgId int64, result tproto.TObject) {

	anyResult, _ := anypb.New(result)
	rpcResult := &tproto.RpcResult{
		ReqMsgId: reqMsgId,
		Result:   anyResult,
	}

	body := tproto.EncodeTObject(rpcResult)

	rawMsg := &tproto.MsgRawData{
		MsgId: nextMessageId(),
		SeqNo: c.generateMessageSeqNo(true),
		Bytes: int32(len(body)),
		Body:  body,
	}

	c.outQueue.AddRpcResultMessage(reqMsgId, rawMsg)
	// cb(rawMsg)
}

func (c *session) sendPushRpcResultToQueue(gatewayId string, reqMsgId int64, result []byte) {
	rawMsg := &tproto.MsgRawData{
		MsgId: nextMessageId(),
		SeqNo: c.generateMessageSeqNo(true),
		Bytes: int32(len(result)),
		Body:  result,
	}
	c.outQueue.AddRpcResultMessage(reqMsgId, rawMsg)
}

func (c *session) sendPushToQueue(ctx context.Context, gatewayId string, pushMsgId int64, pushMsg tproto.TObject) {

	body := tproto.EncodeTObject(pushMsg)

	//if x.GetOffset() > 256 {
	//	gzipPacked := &tproto.TGzipPacked{
	//		PackedData: rawBytes,
	//	}
	//	x2 := tproto.NewEncodeBuf(512)
	//	_ = gzipPacked.Encode(x2)
	//	rawBytes = x2.GetBuf()
	//}

	rawMsg := &tproto.MsgRawData{
		MsgId: nextMessageId(),
		SeqNo: c.generateMessageSeqNo(true),
		Bytes: int32(len(body)),
		Body:  body,
	}
	c.outQueue.AddPushUpdates(pushMsgId, rawMsg)
}

func (c *session) sendRawToQueue(ctx context.Context, gatewayId string, msgId int64, confirm bool, rawMsg tproto.TObject) {
	body := tproto.EncodeTObject(rawMsg)
	rawMsg2 := &tproto.MsgRawData{
		MsgId: nextMessageId(),
		SeqNo: c.generateMessageSeqNo(confirm),
		Bytes: int32(len(body)),
		Body:  body,
	}
	c.outQueue.AddNotifyMessage(msgId, confirm, rawMsg2)
}

func (c *session) sendDirectToGateway(ctx context.Context, gatewayId string, confirm bool, obj tproto.TObject, cb func(sentRaw *tproto.MsgRawData)) (bool, error) {
	if c.connState != StateOnline {
		return false, nil
	}

	body := tproto.EncodeTObject(obj)

	rawMsg := &tproto.MsgRawData{
		MsgId: nextMessageId(),
		SeqNo: c.generateMessageSeqNo(confirm),
		Bytes: int32(len(body)),
		Body:  body,
	}

	var (
		rB  bool
		err error
	)

	if !c.isHttp {
		rB, err = c.sessList.cb.cb.Service.SendDataToGateway(
			ctx,
			gatewayId,
			c.sessList.authId,
			c.sessionId,
			rawMsg)
	} else {
		if ch := c.httpQueue.Pop(); ch != nil {
			rB, err = c.sessList.cb.cb.Service.SendHttpDataToGateway(
				ctx,
				ch,
				c.sessList.authId,
				c.sessionId,
				rawMsg)
		}
	}

	if err != nil {
		logx.WithContext(ctx).Errorf("sendToClient - %v", err)
	}

	if cb != nil {
		cb(rawMsg)
	}

	return rB, err
}

func (c *session) sendRawDirectToGateway(ctx context.Context, gatewayId string, raw *tproto.MsgRawData) (bool, error) {
	if c.connState != StateOnline {
		return false, nil
	}

	var (
		rB  bool
		err error
	)

	if !c.isHttp {
		rB, err = c.sessList.cb.cb.Service.SendDataToGateway(
			ctx,
			gatewayId,
			c.sessList.authId,
			c.sessionId,
			raw)
	} else {
		if ch := c.httpQueue.Pop(); ch != nil {
			rB, err = c.sessList.cb.cb.Service.SendHttpDataToGateway(
				ctx,
				ch,
				c.sessList.authId,
				c.sessionId,
				raw)
		}
	}

	if err != nil {
		logx.WithContext(ctx).Errorf("sendRawDirectToGateway - %v", err)
	}
	return rB, err
}

func (c *session) sendQueueToGateway(ctx context.Context, gatewayId string) {
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

	//for _, m := range listPending {
	//	logx.Debugf("ListPending: %+v - %d - %d", m.msgId, m.msg.MsgId, m.msg.ClassId)
	//}

	if len(listPending) == 1 {
		logx.WithContext(ctx).Infof("sendRawDirectToGateway - pending[0]")
		b, err = c.sendRawDirectToGateway(ctx, gatewayId, listPending[0].msg)
		// log.Debugf("err: %v, b: %v", err, b)
		if err != nil || !b {
			return
		}

		for _, m := range listPending {
			if m.state == NO_NEED_ACK {
				logx.WithContext(ctx).Infof("no_need_ack: %d", m.msgId)
				c.outQueue.Remove(m.msgId)
			} else {
				logx.WithContext(ctx).Infof("pending sent: %d", m.msgId)
				m.sent = sentTime
			}
		}
	} else if len(listPending) > 1 {
		var (
			split = 16
		)
		for i := 0; i < len(listPending)/split; i++ {
			msgContainer := &tproto.TMsgRawDataContainer{
				Messages: make([]*tproto.MsgRawData, 0, split),
			}
			for _, m := range listPending[i*split : (i+1)*split] {
				msgContainer.Messages = append(msgContainer.Messages, m.msg)
			}
			logx.WithContext(ctx).Infof("sendRawDirectToGateway - TLMsgRawDataContainer")
			b, err = c.sendDirectToGateway(ctx, gatewayId, false, msgContainer, func(sentRaw *tproto.MsgRawData) {
			})
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
			msgContainer := &tproto.TMsgRawDataContainer{
				Messages: make([]*tproto.MsgRawData, 0, split),
			}
			for _, m := range listPending[split*(len(listPending)/split):] {
				msgContainer.Messages = append(msgContainer.Messages, m.msg)
			}
			logx.WithContext(ctx).Infof("sendRawDirectToGateway - TLMsgRawDataContainer")
			b, err = c.sendDirectToGateway(ctx, gatewayId, false, msgContainer, func(sentRaw *tproto.MsgRawData) {
			})
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
