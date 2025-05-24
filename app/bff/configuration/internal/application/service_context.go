package application

import (
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/config"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/service"
)

type ServiceContext struct {
	Config config.Config
	*service.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	logicService := service.New(c)
	return &ServiceContext{
		Config:  c,
		Service: logicService,
	}
}
