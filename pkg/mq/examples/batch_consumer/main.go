package main

import (
	"flag"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/commands"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "c.yaml", "the config file")

type Config struct {
	service.ServiceConf
	TestConsumer kafka.KafkaConsumerConf
}

type Server struct {
	// grpcSrv *zrpc.RpcServer
	mq  *kafka.BatchConsumerGroup
	mq2 *kafka.BatchConsumerGroupV2
}

func New() *Server {
	return new(Server)
}

func (s *Server) Initialize() error {
	var c Config
	conf.MustLoad(*configFile, &c)
	logx.SetUp(c.Log)
	logx.Infov(c)

	// s.doBatchConsumer(c)
	s.doBatchConsumerV2(c)

	return nil
}

func (s *Server) RunLoop() {
}

func (s *Server) Destroy() {
	s.mq.Stop()
	// s.grpcSrv.Stop()
}

func (s *Server) doBatchConsumer(c Config) {
	mq := kafka.MustKafkaBatchConsumer(&c.TestConsumer)

	mq.RegisterHandlers(
		func(triggerID string, idList []string) {
			logx.Debug("triggerID: ", triggerID, ", idList: ", len(idList))
		},
		func(value kafka.MsgChannelValue) {
			for _, msg := range value.MsgList {
				// time.Sleep(time.Millisecond)
				logx.Debug("AggregationID: ", value.AggregationID, ", TriggerID: ", value.TriggerID, ", Msg: ", string(msg.MsgData))
			}
		})

	s.mq = mq
	go s.mq.Start()
}

func (s *Server) doBatchConsumerV2(c Config) {
	mq2 := kafka.MustKafkaBatchConsumerV2(&c.TestConsumer, true)

	mq2.RegisterHandler(
		func(channelID int, msg *kafka.MsgConsumerMessage) {
			fmt.Println("channelID: ", channelID, ", TriggerID: ", msg.TriggerID, ", len: ", len(msg.MsgList))
			for _, value := range msg.MsgList {
				_ = value
				// fmt.Println("channelID: ", channelID, ", TriggerID: ", msg.TriggerID(), ", len: ", len(value.MsgData))
				// time.Sleep(time.Millisecond)
				// logx.Debug("AggregationID: ", msg.Key(), ", TriggerID: ", msg.TriggerID(), ", Msg: ", string(value.MsgData))
			}
		})

	s.mq2 = mq2
	s.mq2.Start()
}

func main() {
	commands.Run(New())
}
