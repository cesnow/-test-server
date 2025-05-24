package grpc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"kiyudesign.com/cesnow/light-server/app/status/internal/application"
	"kiyudesign.com/cesnow/light-server/app/status/internal/server/grpc/service"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
)

func New(ctx *application.ServiceContext, c zrpc.RpcServerConf) *zrpc.RpcServer {
	s, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		statuspb.RegisterRpcStatusServer(grpcServer, service.New(ctx))
	})
	logx.Must(err)
	return s
}
