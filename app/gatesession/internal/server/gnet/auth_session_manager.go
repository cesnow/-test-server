package gnet

import (
	"container/list"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

type sessionData struct {
	sessionId  int64
	connIdList *list.List
}

type authSession struct {
	authId      int64
	sessionList map[int64]sessionData
}

type AuthSessionManager struct {
	rw       sync.RWMutex
	sessions map[int64]*authSession
}

func NewAuthSessionManager() *AuthSessionManager {
	return &AuthSessionManager{
		sessions: make(map[int64]*authSession),
	}
}

func (m *AuthSessionManager) AddNewSession(authId int64, sessionId int64, connId int64) (bNew bool) {
	logx.Debugf("addNewSession: auth_key_id: %d, session_id: %d, conn_id: %d",
		authId,
		sessionId,
		connId)

	m.rw.Lock()
	defer m.rw.Unlock()

	if v, ok := m.sessions[authId]; ok {
		var (
			cExisted = false
		)
		if v2, ok2 := v.sessionList[sessionId]; ok2 {
			for e := v2.connIdList.Front(); e != nil; e = e.Next() {
				if e.Value.(int64) == connId {
					cExisted = true
					break
				}
			}
			if !cExisted {
				v2.connIdList.PushBack(connId)
			}
		} else {
			s := sessionData{
				sessionId:  sessionId,
				connIdList: list.New(),
			}
			s.connIdList.PushBack(connId)
			v.sessionList[sessionId] = s
			bNew = true
		}
	} else {
		s := sessionData{
			sessionId:  sessionId,
			connIdList: list.New(),
		}
		s.connIdList.PushBack(connId)

		m.sessions[authId] = &authSession{
			authId: authId,
			sessionList: map[int64]sessionData{
				sessionId: s,
			},
		}
		bNew = true
	}
	return
}

func (m *AuthSessionManager) RemoveSession(authId, sessionId int64, connId int64) (bDeleted bool) {
	logx.Debugf("removeSession: auth_key_id: %d, session_id: %d, conn_id: %d",
		authId,
		sessionId,
		connId)

	m.rw.Lock()
	defer m.rw.Unlock()

	if v, ok := m.sessions[authId]; ok {
		if v2, ok2 := v.sessionList[sessionId]; ok2 {
			for e := v2.connIdList.Front(); e != nil; e = e.Next() {
				if e.Value.(int64) == connId {
					v2.connIdList.Remove(e)
					break
				}
			}
			if v2.connIdList.Len() == 0 {
				delete(v.sessionList, sessionId)
				bDeleted = true
			}
			if len(v.sessionList) == 0 {
				delete(m.sessions, authId)
			}
		}
	}

	return
}

func (m *AuthSessionManager) FoundSessionConnId(authId, sessionId int64) (int64, []int64) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	if v, ok := m.sessions[authId]; ok {
		if v2, ok2 := v.sessionList[sessionId]; ok2 {
			connIdList := make([]int64, 0, v2.connIdList.Len())
			for e := v2.connIdList.Back(); e != nil; e = e.Prev() {
				connIdList = append(connIdList, e.Value.(int64))
			}
			return v.authId, connIdList
		}
	}

	return 0, nil
}
