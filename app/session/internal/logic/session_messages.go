package logic

import (
	"context"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"math"
	"math/rand"
	"reflect"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (c *session) onNewSessionCreated(ctx context.Context, gatewayId string, msgId int64) {
	logx.WithContext(ctx).Infof("onNewSessionCreated - request gatewayId: %s, msgId: %d", gatewayId, msgId)
	newSessionCreated := &tproto.NewSessionCreated{
		FirstMsgId: msgId,
		UniqueId:   rand.Int63(),
	}

	logx.WithContext(ctx).Infof("onNewSessionCreated - reply: {%v}", newSessionCreated)

	if !c.isHttp {
		_, _ = c.sendDirectToGateway(ctx, gatewayId, true, newSessionCreated, func(sentRaw *tproto.MsgRawData) {
			id2 := c.sessList.cb.getNextNotifyId()
			sentMsg := c.outQueue.AddNotifyMessage(id2, true, sentRaw)
			sentMsg.sent = 0
		})
	}
}

func (c *session) onTestSetUser(ctx context.Context, gatewayId string, msgId *inboxMsg, msg *sessionpb.TSession_TTTTT_TestUser) {
	logx.WithContext(ctx).Infof("onTestSetUser - request data: {sess: %s, gatewayId: %s, msg_id: %s, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId.msgId,
		msgId.seqNo,
		msg)

	c.sessList.cb.onSetMainUpdatesSession(ctx, c)
	c.sessList.cb.changeAuthState(ctx, AuthStateNormal, msg.UserId)

	result := tproto.BoolTrue
	c.sendRawToQueue(ctx, gatewayId, msgId.msgId, false, result)
	msgId.state = RECEIVED | NO_NEED_ACK
}

func (c *session) onPing(ctx context.Context, gatewayId string, msgId *inboxMsg, ping *tproto.Ping) {
	logx.WithContext(ctx).Infof("onPing - request data: {sess: %s, gatewayId: %s, msg_id: %s, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId.msgId,
		msgId.seqNo,
		ping)

	pong := &tproto.Pong{
		MsgId:  msgId.msgId,
		PingId: ping.PingId,
	}

	c.sendRawToQueue(ctx, gatewayId, msgId.msgId, false, pong)
	msgId.state = RECEIVED | NO_NEED_ACK

	c.closeDate = time.Now().Unix() + DefaultPingTimeout + PingAddTimeout
}

func (c *session) onDestroySession(ctx context.Context, gatewayId string, msgId *inboxMsg, request *tproto.DestroySession) {
	logx.WithContext(ctx).Infof("onDestroySession - request data: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId.msgId,
		msgId.seqNo,
		request)

	if request.SessionId == c.sessionId {
		logx.WithContext(ctx).Error("the result of this being applied to the current session is undefined.")
		return
	}

	if c.sessList.destroySession(request.GetSessionId()) {
		destroySessionOk := &tproto.DestroySessionOk{
			Session: &tproto.DestroySession{
				SessionId: request.SessionId,
			},
		}
		c.sendRawToQueue(ctx, gatewayId, msgId.msgId, false, destroySessionOk)
	} else {
		destroySessionNone := &tproto.DestroySessionNone{
			Session: &tproto.DestroySession{
				SessionId: request.SessionId,
			},
		}
		c.sendRawToQueue(ctx, gatewayId, msgId.msgId, false, destroySessionNone)
	}

	msgId.state = RECEIVED | NO_NEED_ACK
}

func (c *session) onMsgAck(ctx context.Context, gatewayId string, msgId int64, seqNo int32, request *tproto.TMsgAck) {
	logx.WithContext(ctx).Infof("onMsgAck - request data: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId,
		seqNo,
		request)

	c.outQueue.OnMessagesAck(request.GetMsgIds(), func(inMsgId int64) {
		if inMsgId > math.MaxInt32 {
			c.inQueue.ChangeAckReceived(inMsgId)
			return
		}
	})
}

func (c *session) checkBadMsgNotification(ctx context.Context, gatewayId string, excludeMsgIdToo bool, msg *tproto.TMessage) bool {
	var errorCode int32 = 0

	serverTime := time.Now().Unix()
	clientTime := int64(msg.MsgId / 4294967296.0)

	// Check client time
	if !excludeMsgIdToo {
		if clientTime+60 < serverTime {
			errorCode = MsgIdTooLow
			logx.WithContext(ctx).Errorf("bad server time - {msg_id: %d, clientTime: %d, serverTime: %d}", msg.MsgId, clientTime, serverTime)
		} else if clientTime > serverTime+300 {
			errorCode = MsgIdTooHigh
			logx.WithContext(ctx).Errorf("bad server time - {msg_id: %d, clientTime: %d, serverTime: %d}", msg.MsgId, clientTime, serverTime)
		}
	}
	// Handle error case
	if errorCode != 0 {
		badMsgNotification := &tproto.BadMsgNotification{
			BadMsgId:    msg.MsgId,
			BadMsgSeqno: msg.SeqNo,
			ErrorCode:   errorCode,
		}

		actualObj, _ := msg.Object.UnmarshalNew()

		logx.WithContext(ctx).Error("errorCode - ", errorCode, ", msg: ", reflect.TypeOf(actualObj))

		_, _ = c.sendDirectToGateway(ctx, gatewayId, false, badMsgNotification, func(sentRaw *tproto.MsgRawData) {
			// nothing to do
		})
		return false
	}

	return true
}
