package config

import (
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/zrpc"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
)

type Config struct {
	zrpc.RpcServerConf
	KV               kv.KvConf
	BizServiceClient zrpc.RpcClientConf
	SyncClient       *kafka.KafkaProducerConf
}
