package grpc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/application"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/server/grpc/service"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
)

func New(ctx *application.ServiceContext, c zrpc.RpcServerConf) *zrpc.RpcServer {
	s, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		configurationpb.RegisterRpcConfigurationServer(grpcServer, service.New(ctx))
	})
	logx.Must(err)
	return s
}
