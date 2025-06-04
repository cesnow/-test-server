package authsession

import (
	"context"
	"encoding/json"
	"kiyudesign.com/cesnow/light-server/pkg/transport"
	"math"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

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
