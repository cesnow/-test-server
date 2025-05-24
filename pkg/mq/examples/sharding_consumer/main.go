package main

import (
	"context"
	"flag"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/commands"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "c.yaml", "the config file")

type Config struct {
	service.ServiceConf
	TestConsumer kafka.KafkaShardingConsumerConf
}

type Server struct {
	grpcSrv *zrpc.RpcServer
	mq      *kafka.ShardingConsumerGroup
}

func New() *Server {
	return new(Server)
}

func (s *Server) Initialize() error {
	var c Config
	conf.MustLoad(*configFile, &c)
	logx.SetUp(c.Log)

	logx.Infov(c)
	mq := kafka.MustShardingConsumerGroup(&c.TestConsumer)

	mq.RegisterHandler(func(ctx context.Context, method, key string, value []byte) {
		fmt.Println("key: ", key, ", value: ", string(value))
	})

	s.mq = mq
	go s.mq.Start()

	return nil
}

func (s *Server) RunLoop() {
}

func (s *Server) Destroy() {
	s.mq.Stop()
	s.grpcSrv.Stop()
}

func main() {
	commands.Run(New())
}
