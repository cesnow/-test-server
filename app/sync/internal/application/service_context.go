package application

import (
	"kiyudesign.com/cesnow/light-server/app/sync/internal/config"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/service"
)

type ServiceContext struct {
	Config config.Config
	*service.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		Service: service.New(c),
	}
}
