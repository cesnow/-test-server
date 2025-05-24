package configuration_helper

import (
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/application"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/config"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/server/grpc/service"
)

type (
	Config = config.Config
)

func New(c Config) *service.Service {
	return service.New(application.NewServiceContext(c))
}
