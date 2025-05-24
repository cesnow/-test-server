package server

import (
	"flag"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/application"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/config"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/gnet"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/grpc"
)

var (
	configFile = flag.String("f", "/etc/ice/gateway.yaml", "the config file")
)

type Server struct {
	grpcSrv *zrpc.RpcServer
	server  *gnet.Server
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)

	ctx := application.NewServiceContext(c)
	s.server = gnet.New(ctx, c)
	s.grpcSrv = grpc.New(ctx, c.RpcServerConf, s.server)

	go func() {
		s.grpcSrv.Start()
	}()

	return nil
}

func (s *Server) RunLoop() {
	// s.server.Serve()
	//if err := s.server.Serve(); err != nil {
	//	logx.Errorf("run server error: %v, quit...", err)
	//	commands.GSignal <- syscall.SIGQUIT
	//}
}

func (s *Server) Destroy() {
	s.grpcSrv.Stop()
	s.server.Close()
}
