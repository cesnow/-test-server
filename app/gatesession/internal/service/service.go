package service

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/config"
	"kiyudesign.com/cesnow/light-server/pkg/net/ip"
)

type Service struct {
	GatewayId string
	KV        kv.Store
}

func New(c config.Config) (s *Service) {
	s = new(Service)
	s.GatewayId = ip.FigureOutListenOn(c.ListenOn)
	s.KV = kv.NewStore(c.KV)
	logx.Infof("MyGatewayId: %s", s.GatewayId)
	return s
}
