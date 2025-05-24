package logic

import (
	"kiyudesign.com/cesnow/light-server/app/session/internal/service"
	"strconv"
	"sync"
)

type MainAuthWrapperManager struct {
	mu      sync.Mutex
	authMgr map[int64]*MainAuthWrapper
	*service.Service
}

func NewMainAuthWrapperManager(d *service.Service) *MainAuthWrapperManager {
	return &MainAuthWrapperManager{
		authMgr: make(map[int64]*MainAuthWrapper),
		Service: d,
	}
}

func (m *MainAuthWrapperManager) GetMainAuthWrapper(mainAuthId int64) *MainAuthWrapper {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.authMgr[mainAuthId]
	if ok {
		return v
	}

	return nil
}

func (m *MainAuthWrapperManager) DeleteByAuthId(mainAuthId int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.authMgr, mainAuthId)
}

func (m *MainAuthWrapperManager) AllocMainAuthWrapper(authId int64, newMainAuth func(authId int64) *MainAuthWrapper) *MainAuthWrapper {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.authMgr[authId]
	if ok {
		return v
	} else {
		mainAuth := newMainAuth(authId)
		m.authMgr[authId] = mainAuth
		return mainAuth
	}
}

func (m *MainAuthWrapperManager) OnShardingCB(sharding *service.RpcShardingManager, oldList, addList []string, removeList []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.authMgr {
		if !sharding.ShardingVIsListenOn(strconv.FormatInt(k, 10)) {
			delete(m.authMgr, k)
			v.Stop()
		}
	}
}
