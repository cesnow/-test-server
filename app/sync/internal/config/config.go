package config

import (
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/zrpc"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
)

type Config struct {
	zrpc.RpcServerConf
	KV            kv.KvConf
	Routine       Routine
	SyncConsumer  kafka.KafkaConsumerConf
	SessionClient zrpc.RpcClientConf
	StatusClient  zrpc.RpcClientConf
}

type Routine struct {
	Size uint64
	Chan uint64
}
