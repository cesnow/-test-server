package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
	"kiyudesign.com/cesnow/light-server/pkg/stores/mon"
)

type Config struct {
	zrpc.RpcServerConf
	SyncClient *kafka.KafkaProducerConf

	Cache cache.CacheConf
	Mongo mon.Config
}
