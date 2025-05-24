package server

import (
	"flag"
	"kiyudesign.com/cesnow/light-server/app/session/internal/application"
	"kiyudesign.com/cesnow/light-server/app/session/internal/config"
	"kiyudesign.com/cesnow/light-server/app/session/internal/server/grpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "/etc/ice/session.yaml", "the config file")

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
		s.grpcSrv.Start()
	}()
	return nil
}

func (s *Server) RunLoop() {
}

func (s *Server) Destroy() {
	s.grpcSrv.Stop()
}
