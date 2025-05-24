package service

import (
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/svc"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"
)

type Service struct {
	svcCtx *svc.ServiceContext
	gatewaypb.RpcGatewayServer
}

func New(ctx *svc.ServiceContext, srv gatewaypb.RpcGatewayServer) *Service {
	return &Service{
		svcCtx:           ctx,
		RpcGatewayServer: srv,
	}
}
