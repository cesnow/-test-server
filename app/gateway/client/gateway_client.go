package gateway_client

import (
	"context"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"

	"github.com/zeromicro/go-zero/zrpc"
)

type GatewayClient interface {
	GatewaySendDataToGateway(ctx context.Context, in *gatewaypb.TGatewaySendDataToGateway) (*tproto.Bool, error)
}

type defaultGatewayClient struct {
	cli zrpc.Client
}

func NewGatewayClient(cli zrpc.Client) GatewayClient {
	return &defaultGatewayClient{
		cli: cli,
	}
}

func (m *defaultGatewayClient) GatewaySendDataToGateway(ctx context.Context, in *gatewaypb.TGatewaySendDataToGateway) (*tproto.Bool, error) {
	client := gatewaypb.NewRpcGatewayClient(m.cli.Conn())
	return client.GatewaySendDataToGateway(ctx, in)
}
