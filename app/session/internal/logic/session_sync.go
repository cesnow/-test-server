package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
)

func (c *session) onSyncData(ctx context.Context, obj tproto.TObject) {
	if obj != nil {
		logx.WithContext(ctx).Infof("session]]>> - session: %s, syncData: %s", c, obj)
	} else {
		logx.WithContext(ctx).Infof("session]]>> - session: %s, syncData: nil", c)
	}

	gatewayId := c.getGatewayId()

	pushMessageId := c.sessList.cb.getNextPushId()
	c.sendPushToQueue(ctx, gatewayId, pushMessageId, obj)

	if c.sessionOnline() {
		if gatewayId == "" {
			logx.WithContext(ctx).Errorf("gatewayId is empty, send delay...")
		} else {
			c.sendQueueToGateway(ctx, gatewayId)
		}
	}
}

func (c *session) onSyncRpcResultData(ctx context.Context, reqMsgId int64, data []byte) {
	logx.WithContext(ctx).Debugf("onSyncRpcResultData]]>> - %s", data)

	c.pendingQueue.Remove(reqMsgId)
	gatewayId := c.getGatewayId()
	c.sendPushRpcResultToQueue(gatewayId, reqMsgId, data)
}

func (c *session) onSyncSessionData(ctx context.Context, obj tproto.TObject) {
	gatewayId := c.getGatewayId()
	pushMsgId := c.sessList.cb.getNextPushId()

	c.sendPushToQueue(ctx, gatewayId, pushMsgId, obj)
	c.sendQueueToGateway(ctx, gatewayId)
}
