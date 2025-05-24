package authsession

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

func (c *session) onSyncData(ctx context.Context, obj []byte) {
	if obj != nil {
		logx.WithContext(ctx).Infof("session]]>> - session: %s, syncData: %s", c, obj)
	} else {
		logx.WithContext(ctx).Infof("session]]>> - session: %s, syncData: nil", c)
	}

	gatewayId := c.getGatewayId()

	pushMessageId := c.sessList.cb.getNextPushId()
	//c.sendPushToQueue(ctx, gatewayId, pushMessageId, obj)

	c.sendRawToQueue(ctx, gatewayId, pushMessageId, false, obj)

	if c.sessionOnline() {
		if gatewayId == "" {
			logx.WithContext(ctx).Errorf("gatewayId is empty, send delay...")
		} else {
			c.sendQueueToClient(ctx, gatewayId)
		}
	}
}
