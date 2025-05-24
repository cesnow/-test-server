package service

import (
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/zrpc"
	statusclient "kiyudesign.com/cesnow/light-server/app/status/client"
	syncclient "kiyudesign.com/cesnow/light-server/app/sync/client"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/config"
)

type Service struct {
	kv             kv.Store
	conf           *config.Config
	sessionServers map[string]*Session
	PushClient     syncclient.SyncClient
	statusclient.StatusClient
}

func New(c config.Config) *Service {
	d := &Service{
		kv:             kv.NewStore(c.KV),
		conf:           &c,
		sessionServers: make(map[string]*Session),
		StatusClient:   statusclient.NewStatusClient(zrpc.MustNewClient(c.StatusClient)),
	}

	go d.watch(c.SessionClient)
	return d
}
