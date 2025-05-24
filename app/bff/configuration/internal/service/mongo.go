package service

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/persistence/repository/mongodb"
	"kiyudesign.com/cesnow/light-server/pkg/stores/mon"
)

// ---

type MongoDao struct {
	*mon.DB
	AuthsModel *mongodb.AuthsModel
}

func newMongoDao(db *mon.DB, cache cache.CacheConf) *MongoDao {
	return &MongoDao{
		db,
		mongodb.NewTest(db, cache),
	}
}
