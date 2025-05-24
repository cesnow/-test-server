package server

import (
	"flag"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration"
	"kiyudesign.com/cesnow/light-server/app/bff/proxy/internal/config"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
)

var configFile = flag.String("f", "/etc/bff.yaml", "the config file")

type Server struct {
	grpcSrv *zrpc.RpcServer
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)

	s.grpcSrv = zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// configuration_helper
		configurationpb.RegisterRpcConfigurationServer(
			grpcServer,
			configuration_helper.New(configuration_helper.Config{
				RpcServerConf: c.RpcServerConf,
				SyncClient:    c.SyncClient,
			}))
	})

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
