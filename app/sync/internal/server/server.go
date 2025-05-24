package server

import (
	"flag"
	"github.com/zeromicro/go-zero/zrpc"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/application"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/config"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/server/grpc"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/server/mq"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/sync.yaml", "the config file")

type Server struct {
	grpcSrv *zrpc.RpcServer
	mq      *kafka.ConsumerGroup
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)

	if err := logx.SetUp(c.Log); err != nil {
		return err
	}

	ctx := application.NewServiceContext(c)
	s.grpcSrv = grpc.New(ctx, c.RpcServerConf)
	s.mq = mq.New(ctx, c.SyncConsumer)

	return nil
}

func (s *Server) RunLoop() {
	go s.grpcSrv.Start()
	go s.mq.Start()
}

func (s *Server) Destroy() {
	// s.grpcSrv.Stop()
	s.mq.Stop()
}
