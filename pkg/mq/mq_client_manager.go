package kafka

import (
	"io"

	"github.com/zeromicro/go-zero/core/syncx"
)

var (
	clientManager = syncx.NewResourceManager()
)

func GetCachedMQClient(c *KafkaProducerConf) *Producer {
	var (
		val io.Closer
		err error
	)
	val, err = clientManager.GetResource(c.Topic, func() (io.Closer, error) {
		cli := MustKafkaProducer(c)
		return cli, nil
	})
	if err != nil {
		panic(err)
	}

	return val.(*Producer)
}
