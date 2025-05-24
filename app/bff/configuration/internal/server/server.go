package server

import (
	"flag"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/application"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/config"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/server/grpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/configuration.yaml", "the config file")

type Server struct {
	grpcSrv *zrpc.RpcServer
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)
	ctx := application.NewServiceContext(c)
	s.grpcSrv = grpc.New(ctx, c.RpcServerConf)

	go func() {
		go s.grpcSrv.Start()
	}()
	return nil
}

func (s *Server) RunLoop() {
}

func (s *Server) Destroy() {
	s.grpcSrv.Stop()
}
