package grpc

import (
	"kiyudesign.com/cesnow/light-server/app/session/internal/application"
	"kiyudesign.com/cesnow/light-server/app/session/internal/server/grpc/service"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func New(ctx *application.ServiceContext, c zrpc.RpcServerConf) *zrpc.RpcServer {
	s, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		sessionpb.RegisterRpcSessionServer(grpcServer, service.New(ctx))
	})
	logx.Must(err)
	return s
}
