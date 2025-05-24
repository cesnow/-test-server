package svc

import (
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/config"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/service"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/service/authsession"
)

type ServiceContext struct {
	Config config.Config
	*service.Service
	MainAuthMgr *authsession.MainAuthWrapperManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	s := service.New(c)
	mainAuthMgr := authsession.NewMainAuthWrapperManager(s)
	return &ServiceContext{
		Config:      c,
		Service:     s,
		MainAuthMgr: mainAuthMgr,
	}
}
