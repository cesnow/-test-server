package application

import (
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/client"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/config"
	"kiyudesign.com/cesnow/light-server/pkg/net/ip"
)

type ServiceContext struct {
	Config config.Config
	*client.Client

	GatewayId string
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		Client:    client.New(c),
		GatewayId: ip.FigureOutListenOn(c.ListenOn),
	}
}
