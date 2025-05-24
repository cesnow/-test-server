package service

import (
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/application"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"
)

type Service struct {
	svcCtx *application.ServiceContext
	gatewaypb.RpcGatewayServer
}

func New(ctx *application.ServiceContext, srv gatewaypb.RpcGatewayServer) *Service {
	return &Service{
		svcCtx:           ctx,
		RpcGatewayServer: srv,
	}
}
