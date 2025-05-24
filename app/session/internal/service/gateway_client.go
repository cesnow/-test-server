package service

import (
	"context"
	gatewayclient "kiyudesign.com/cesnow/light-server/app/gateway/client"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type Gateway struct {
	serverId string
	client   gatewayclient.GatewayClient
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewGateway(c zrpc.RpcClientConf) (*Gateway, error) {
	g := &Gateway{
		serverId: c.Endpoints[0],
	}

	cli, err := zrpc.NewClient(
		c,
		zrpc.WithDialOption(grpc.WithReadBufferSize(16*1024*1024)),
		zrpc.WithDialOption(grpc.WithWriteBufferSize(16*1024*1024)))
	if err != nil {
		logx.Errorf("watchComet NewClient(%+v) error(%v)", c, err)
		return nil, err
	}
	g.client = gatewayclient.NewGatewayClient(cli)
	g.ctx, g.cancel = context.WithCancel(context.Background())

	return g, nil
}

func (c *Gateway) Close() (err error) {
	c.cancel()
	return
}

func (c *Gateway) SendDataToGate(ctx context.Context, authId, sessionId int64, payload []byte) (b bool, err error) {
	var (
		res *tproto.Bool
	)

	res, err = c.client.GatewaySendDataToGateway(ctx, &gatewaypb.TGatewaySendDataToGateway{
		AuthId:    authId,
		SessionId: sessionId,
		Payload:   payload,
	})

	if err != nil {
		logx.Errorf("sendDataToGate error: %v", err)
		b = false
		return
	}

	b = tproto.FromBool(res)
	return
}
