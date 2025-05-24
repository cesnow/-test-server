package grpc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/application"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/server/grpc/service"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
)

func New(ctx *application.ServiceContext, c zrpc.RpcServerConf) *zrpc.RpcServer {
	s, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		syncpb.RegisterRpcSyncServer(grpcServer, service.New(ctx))
		_ = service.New(ctx)
	})
	logx.Must(err)
	return s
}
