package authsession

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/threading"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
	"kiyudesign.com/cesnow/light-server/pkg/transport"
	"math"
	"strconv"
	"sync"
	"time"
)

const (
	AuthStateUnknown      = 0
	AuthStateNew          = 1
	AuthStateUnauthorized = 2
	AuthStateNormal       = 3
	AuthStateLogout       = 4
	AuthStateDeleted      = 5
)

type SessionList struct {
	authId   int64
	state    int
	sessions map[int64]*session
	cb       *MainAuthWrapper
}

func newSessionList(cb *MainAuthWrapper) *SessionList {
	return &SessionList{
		authId:   0,
		state:    0,
		sessions: make(map[int64]*session),
		cb:       cb,
	}
}

func (s *SessionList) Reset(authId int64) (lastAuthId int64) {
	lastAuthId = s.authId

	s.authId = authId
	s.state = 0
	s.sessions = make(map[int64]*session)

	return
}

func (s *SessionList) destroySession(sessionId int64) bool {
	if sess, ok := s.sessions[sessionId]; ok {
		logx.Infof("session]]>> destroySession: %d", sess.sessionId)
		delete(s.sessions, sessionId)
	}
	return true
}

func (s *SessionList) changeAuthState(state int) {
	s.state = state
}

type MainAuthWrapper struct {
	authId             int64
	state              int
	AuthUserId         int64
	mainAuth           *SessionList
	mainUpdatesSession *session
	closeChan          chan struct{}
	sessionDataChan    chan any
	finish             sync.WaitGroup
	running            *syncx.AtomicBool
	onlineExpired      int64
	clientType         int
	nextNotifyId       int64
	nextPushId         int64
	cb                 *MainAuthWrapperManager
	expiredTime        int64
	sendCb             func(ctx context.Context, gatewayId string, authId int64, sessionId int64, data *transport.TMsgRawData) (bool, error)
	eventHandler       func(name string, input interface{}) (interface{}, error)
}

func NewMainAuthWrapper(mainAuthId int64, authUserId int64, state int, cb *MainAuthWrapperManager, sendCb func(ctx context.Context, gatewayId string, authId int64, sessionId int64, data *transport.TMsgRawData) (bool, error), eventHandler func(name string, input interface{}) (interface{}, error)) *MainAuthWrapper {
	mainAuth := &MainAuthWrapper{
		authId:             mainAuthId,
		state:              state,
		AuthUserId:         authUserId,
		mainAuth:           nil,
		mainUpdatesSession: nil,
		clientType:         0,
		nextPushId:         0,
		nextNotifyId:       math.MaxInt32,
		closeChan:          make(chan struct{}),
		sessionDataChan:    make(chan any, 1024),
		finish:             sync.WaitGroup{},
		running:            syncx.NewAtomicBool(),
		cb:                 cb,
		expiredTime:        120,
		sendCb:             sendCb,
		eventHandler:       eventHandler,
	}
	mainAuth.mainAuth = newSessionList(mainAuth)

	mainAuth.Start()
	return mainAuth
}

func (m *MainAuthWrapper) changeAuthState(ctx context.Context, state int, stateData int64) {
	m.state = state

	switch state {
	case AuthStateUnknown:
		m.cb.DeleteByAuthId(m.authId)
		m.Stop()
	case AuthStateLogout:
		m.cb.DeleteByAuthId(m.authId)
		m.Stop()
	case AuthStateDeleted:
		m.cb.DeleteByAuthId(m.authId)
		m.Stop()
	case AuthStateNormal:
		m.AuthUserId = stateData
	default:
		m.AuthUserId = 0
	}
}

func (m *MainAuthWrapper) resetAuth(authId int64) (lastAuthId int64) {
	lastAuthId = m.mainAuth.Reset(authId)
	m.mainUpdatesSession = nil
	return
}

func (m *MainAuthWrapper) setOnline(ctx context.Context) {
	date := time.Now().Unix()
	if (m.onlineExpired == 0 || date > m.onlineExpired-PingAddTimeout) && m.AuthUserId != 0 {
		logx.WithContext(ctx).Debugf("[DEBUG] setOnline - set online: (date: %d, userId:%d, onlineExpired: %d, authId: %d)",
			date,
			m.AuthUserId,
			m.onlineExpired,
			m.authId)

		var (
			userK = GetOnlineUserKey(m.AuthUserId)
			sess  = &statuspb.SessionEntry{
				UserId:  m.AuthUserId,
				AuthId:  m.authId,
				Gateway: m.cb.Service.GatewayId,
				Expired: date + m.expiredTime,
				Client:  "",
			}
		)

		sessData, _ := jsonx.Marshal(sess)
		err := m.cb.KV.HsetCtx(
			ctx,
			userK,
			strconv.FormatInt(sess.AuthId, 10),
			string(sessData))
		if err != nil {
			logx.Errorf("auth_wrapper.setSessionOnline(%d) error(%v)", m.AuthUserId, err)
			return
		}

		err = m.cb.KV.ExpireCtx(
			ctx,
			userK,
			int(m.expiredTime))
		if err != nil {
			logx.Errorf("auth_wrapper.setSessionOnline(%d) error(%v)", m.AuthUserId, err)
			return
		}

		m.onlineExpired = date + m.expiredTime
	} else {
		//logx.WithContext(ctx).Debugf("[DEBUG] setOnline - not set online: (date: %d, onlineExpired: %d, AuthUserId: %d)",
		//	date,
		//	m.onlineExpired,
		//	m.AuthUserId)
	}
}

func (m *MainAuthWrapper) trySetOffline(ctx context.Context) {
	if m.mainUpdatesSession != nil && m.mainUpdatesSession.sessionOnline() {
		return
	}

	if m.AuthUserId > 0 {
		logx.WithContext(ctx).Infof("[authSessions] offline: %s", m)

		_, err := m.cb.Service.KV.HdelCtx(
			ctx,
			GetOnlineUserKey(m.AuthUserId),
			strconv.FormatInt(m.authId, 10))
		if err != nil {
			logx.Errorf("auth_wrapper.setSessionOffline(%d) error(%v)", m.AuthUserId, err)
			return
		}

	}
	m.onlineExpired = 0
}

func (m *MainAuthWrapper) delOnline(ctx context.Context) {
	if m.AuthUserId > 0 {
		logx.Infof("[authSessions] delOnline: %s", m)

		_, err := m.cb.Service.KV.HdelCtx(
			ctx,
			GetOnlineUserKey(m.AuthUserId),
			strconv.FormatInt(m.authId, 10))
		if err != nil {
			logx.Errorf("status.setSessionOffline(%s) error(%v)", m.authId, err)
			return
		}
	}
	m.onlineExpired = 0
}

func (m *MainAuthWrapper) getNextNotifyId() (id int64) {
	id = m.nextNotifyId
	m.nextNotifyId--
	return
}

func (m *MainAuthWrapper) getNextPushId() (id int64) {
	id = m.nextPushId
	m.nextPushId++
	return
}

func (m *MainAuthWrapper) onSetMainUpdatesSession(ctx context.Context, sess *session) {
	if m.mainUpdatesSession == nil || m.mainUpdatesSession.sessionId != sess.sessionId {
		m.mainUpdatesSession = sess
	}
	m.setOnline(ctx)
}

func (m *MainAuthWrapper) getSessionList() (sList *SessionList) {
	sList = m.mainAuth
	return
}

func (m *MainAuthWrapper) getSessionListById(authId int64) (sList *SessionList) {
	if authId == m.mainAuth.authId {
		sList = m.mainAuth
	}

	return
}

func (m *MainAuthWrapper) String() string {
	return fmt.Sprintf("{auth_key_id: %d, user_id: %d}", m.authId, m.AuthUserId)
}

func (m *MainAuthWrapper) Start() {
	m.running.Set(true)
	m.finish.Add(1)
	go m.runLoop()
}

func (m *MainAuthWrapper) Stop() {
	m.running.Set(false)
}

func (m *MainAuthWrapper) runLoop() {
	defer func() {
		if m.mainUpdatesSession != nil && m.mainUpdatesSession.sessionOnline() {
			m.delOnline(context.Background())
		}
		m.finish.Done()
		close(m.closeChan)
		close(m.sessionDataChan)
		m.finish.Wait()
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for m.running.True() {
		select {
		case <-m.closeChan:
			// log.Info("runLoop -> To Close ", this.String())
			return

		case sessionMsg := <-m.sessionDataChan:
			switch ctxData := sessionMsg.(type) {
			case *sessionDataCtx:
				threading.RunSafe(func() {
					// received data
					m.onSessionData(ctxData.ctx, &ctxData.sessionData)
				})
			case *syncDataCtx:
				threading.RunSafe(func() {
					m.onSyncData(ctxData.ctx, &ctxData.syncData)
				})
			case *connDataCtx:
				threading.RunSafe(func() {
					if ctxData.isNew {
						m.onSessionNew(ctxData.ctx, &ctxData.connData)
					} else {
						m.onSessionClosed(ctxData.ctx, &ctxData.connData)
					}
				})
			default:
				logx.Errorf("receive invalid type msg")
			}
		case <-ticker.C:
			threading.RunSafe(func() {
				m.onTimer(context.Background())
			})
		}
	}

	logx.Infof("%s -> quit runLoop...", m)
}

func (m *MainAuthWrapper) onTimer(ctx context.Context) {
	for _, sess := range m.mainAuth.sessions {
		sess.onTimer(ctx)
	}

	if m.mainUpdatesSession != nil && m.mainUpdatesSession.sessionOnline() {
		m.setOnline(ctx)
	}

	for _, sess := range m.mainAuth.sessions {
		if !sess.sessionClosed() {
			return
		}
	}

	m.cb.DeleteByAuthId(m.authId)
	m.Stop()
}

func (m *MainAuthWrapper) NewSession(ctx context.Context, authId int64, gatewayId string, sessionId int64) error {
	cData := &connDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		connData: connData{
			authId:    authId,
			isNew:     true,
			gatewayId: gatewayId,
			sessionId: sessionId,
		},
	}

	select {
	case m.sessionDataChan <- cData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) SessionDataArrived(ctx context.Context, kId int64, gatewayId, clientIp string, sessionId int64, buf []byte) error {
	sData := &sessionDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		sessionData: sessionData{
			authId:    kId,
			gatewayId: gatewayId,
			clientIp:  clientIp,
			sessionId: sessionId,
			buf:       buf,
		},
	}

	select {
	case m.sessionDataChan <- sData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) CloseSession(ctx context.Context, kId int64, gatewayId string, sessionId int64) error {
	cData := &connDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		connData: connData{
			authId:    kId,
			isNew:     false,
			gatewayId: gatewayId,
			sessionId: sessionId,
		},
	}

	select {
	case m.sessionDataChan <- cData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) SyncDataArrived(ctx context.Context, clientMsgId int64, updates []byte) error {
	sData := &syncDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		syncData: syncData{
			data: &messageData{
				clientMsgId: clientMsgId,
				obj:         updates,
			},
		},
	}

	select {
	case m.sessionDataChan <- sData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) onSessionNew(ctx context.Context, connMsg *connData) {
	sList := m.getSessionList()

	if sList.authId == 0 {
		m.resetAuth(connMsg.authId)
	} else if sList.authId != connMsg.authId {
		m.resetAuth(connMsg.authId)
	}

	sess, ok := sList.sessions[connMsg.sessionId]
	if !ok {
		logx.WithContext(ctx).Infof("onSessionNew - newSession(%d), conn: %s", m.authId, connMsg)
		sess = newSession(connMsg.sessionId, sList)
		sList.sessions[connMsg.sessionId] = sess
	} else {
		sess.sessionState = SessionStateNew
		logx.WithContext(ctx).Infof("onSessionNew - session(%d) found, conn: %s", m.authId, connMsg)
	}

	sess.onSessionConnNew(ctx, connMsg.gatewayId)
}

func (m *MainAuthWrapper) onSessionData(ctx context.Context, sessionMsg *sessionData) {
	sList := m.getSessionList()

	if sList.authId == 0 {
		m.resetAuth(sessionMsg.authId)
	} else if sList.authId != sessionMsg.authId {
		m.resetAuth(sessionMsg.authId)
	}

	logx.Infof("onSessionData - data: {sessions: %s, gate_id: %s}", m, sessionMsg.gatewayId)
	var tMsg *transport.TMsgRawData
	err := msgpack.Unmarshal(sessionMsg.buf, &tMsg)
	if err != nil {
		logx.WithContext(ctx).Errorf("onSessionData - error: {%s}, data: {sessions: %s, gate_id: %d}", err, m, sessionMsg.gatewayId)
		return
	}

	sess, ok := sList.sessions[sessionMsg.sessionId]
	if !ok {
		sess = newSession(sessionMsg.sessionId, sList)
		sList.sessions[sessionMsg.sessionId] = sess
	}

	sess.onSessionConnNew(ctx, sessionMsg.gatewayId)
	sess.onSessionMessageData(ctx, sessionMsg.gatewayId, sessionMsg.clientIp, tMsg)
}

func (m *MainAuthWrapper) onSessionClosed(ctx context.Context, connMsg *connData) {
	sList := m.getSessionList()

	if sess, ok := sList.sessions[connMsg.sessionId]; !ok {
		logx.WithContext(ctx).Errorf("onSessionClosed - session conn closed -  conn: %s", connMsg)
	} else {
		logx.WithContext(ctx).Infof("onSessionClosed - conn: %s, sess: %s", connMsg, sess)
		sess.onSessionConnClose(ctx, connMsg.gatewayId)
	}
}

func (m *MainAuthWrapper) onSyncData(ctx context.Context, syncMsg *syncData) {
	logx.WithContext(ctx).Debugf("authSessions - %s", hex.EncodeToString(syncMsg.data.obj))

	if m.mainUpdatesSession != nil {
		m.mainUpdatesSession.onSyncData(ctx, syncMsg.data.clientMsgId, "", syncMsg.data.obj)
	} else {
		logx.WithContext(ctx).Errorf("authSessions - no mainUpdatesSession")
	}
}
