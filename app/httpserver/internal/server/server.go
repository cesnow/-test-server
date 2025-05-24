package server

import (
	"flag"
	"github.com/zeromicro/go-zero/zrpc"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/config"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/server/grpc"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/server/http"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/httpserver.yaml", "the config file")

type Server struct {
	httpSrv *http.Server
	grpcSrv *zrpc.RpcServer // for t get changes data from grpc
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)
	ctx := svc.NewServiceContext(c)

	s.httpSrv = http.New(ctx, c.Http)
	s.grpcSrv = grpc.New(ctx, c.RpcServerConf, s.httpSrv)

	go func() {
		s.grpcSrv.Start()
	}()

	return nil
}

func (s *Server) RunLoop() {
}

func (s *Server) Destroy() {
}
