package grpc

import (
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/application"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/grpc/service"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func New(svcCtx *application.ServiceContext, c zrpc.RpcServerConf, srv gatewaypb.RpcGatewayServer) *zrpc.RpcServer {
	s, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		gatewaypb.RegisterRpcGatewayServer(grpcServer, service.New(svcCtx, srv))
	})
	s.AddOptions(
		grpc.WriteBufferSize(16*1024*1024),
		grpc.ReadBufferSize(16*1024*1024))

	logx.Must(err)
	return s
}
