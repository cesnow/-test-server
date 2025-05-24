package service

import (
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/config"
	"kiyudesign.com/cesnow/light-server/pkg/net/ip"
)

type Service struct {
	GatewayId string
}

func New(c config.Config) (s *Service) {
	s = new(Service)
	s.GatewayId = ip.FigureOutListenOn(c.ListenOn)
	logx.Infof("MyGatewayId: %s", s.GatewayId)
	return s
}
