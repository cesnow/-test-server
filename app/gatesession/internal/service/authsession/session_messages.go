package authsession

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
	"kiyudesign.com/cesnow/light-server/pkg/transport"
	"math"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestResponse struct {
	Msg string `json:"msg" msgpack:"msg"`
}

func (c *session) onEventRequest(ctx context.Context, gatewayId, clientIp string, inMsg *inboxMsg, event string, data []byte) bool {
	logx.WithContext(ctx).Infof("onEventRequest - request data: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, event: %s, data: %v}",
		c,
		gatewayId,
		inMsg.msgId,
		inMsg.seqNo,
		event, hex.EncodeToString(data))

	switch c.sessList.cb.state {
	case AuthStateNormal:
		// state is ok
	default:
		// TODO: checking without login
	}

	//inMsg.state = RECEIVED | DATA_PROCESSING
	inMsg.state = RECEIVED | NO_NEED_ACK

	var x any
	_ = msgpack.Unmarshal(data, &x)
	logx.WithContext(ctx).Infof("onEventRequest - request data: {data: %+v}", x)

	res := TestResponse{Msg: "hello"}
	rData, _ := msgpack.Marshal(res)

	//ctx:       contextx.ValueOnlyFrom(ctx),
	//	sessList:  c.sessList,
	//		sessionId: c.sessionId,
	//		clientIp:  clientIp,
	//		reqMsgId:  msgId.msgId,
	//		reqMsg:    query,

	c.sendRawToQueue(ctx, gatewayId, inMsg.msgId, false, event, rData)

	return true
}

func (c *session) onMsgAck(ctx context.Context, gatewayId string, msgId int64, msgIds []int64) {
	logx.WithContext(ctx).Infof("onMsgAck - request data: {sess: %s, gatewayId: %s, msg_id: %d, request: {%v}}",
		c,
		gatewayId,
		msgId,
		msgIds)

	c.outQueue.OnMessagesAck(msgIds, func(inMsgId int64) {
		if inMsgId > math.MaxInt32 {
			c.inQueue.ChangeAckReceived(inMsgId)
			return
		}
	})
}

func (c *session) checkBadMsgNotification(ctx context.Context, gatewayId string, excludeMsgIdToo bool, msg *transport.TMsgRawData) bool {
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
		badMsgNotification := map[string]any{
			"badMsgId": msg.MsgId,
			//"badMsgSeqNo": msg.SeqNo,
			"ErrorCode": errorCode,
		}

		logx.WithContext(ctx).Error("errorCode - ", errorCode, ", event: ", msg.Event)

		badMsgNotificationJson, _ := json.Marshal(badMsgNotification)

		_, _ = c.sendDirectToGateway(ctx, gatewayId, false, badMsgNotificationJson)
		return false
	}

	return true
}
