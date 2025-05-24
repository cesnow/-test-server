package service

import (
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/config"
	syncclient "kiyudesign.com/cesnow/light-server/app/sync/client"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
	"kiyudesign.com/cesnow/light-server/pkg/stores/mon"
)

type Service struct {
	syncclient.SyncClient
	*MongoDao
}

func New(c config.Config) *Service {

	mongodb := mon.MustNewMongo(c.Mongo)

	return &Service{
		SyncClient: syncclient.NewSyncMqClient(kafka.MustKafkaProducer(c.SyncClient)),
		MongoDao:   newMongoDao(mongodb, c.Cache),
	}
}
