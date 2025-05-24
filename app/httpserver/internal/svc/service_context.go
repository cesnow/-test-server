package svc

import (
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/config"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/service"
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
