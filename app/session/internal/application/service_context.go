package application

import (
	"kiyudesign.com/cesnow/light-server/app/session/internal/config"
	"kiyudesign.com/cesnow/light-server/app/session/internal/logic"
	"kiyudesign.com/cesnow/light-server/app/session/internal/service"
)

type ServiceContext struct {
	Config      config.Config
	MainAuthMgr *logic.MainAuthWrapperManager
	*service.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	d := service.New(c)
	mainAuthMgr := logic.NewMainAuthWrapperManager(d)
	d.RpcShardingManager.RegisterCB(mainAuthMgr.OnShardingCB)
	d.RpcShardingManager.Start()

	return &ServiceContext{
		Config:      c,
		MainAuthMgr: mainAuthMgr,
		Service:     d,
	}
}
