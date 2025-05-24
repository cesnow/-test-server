package logic

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/threading"
	"google.golang.org/protobuf/types/known/anypb"
	"kiyudesign.com/cesnow/light-server/app/session/internal/service"
	"kiyudesign.com/cesnow/light-server/pkg/queue2"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
	"math"
	"reflect"
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
	authId               int64
	state                int
	AuthUserId           int64
	mainAuth             *SessionList
	mainUpdatesSession   *session
	closeChan            chan struct{}
	sessionDataChan      chan any
	rpcDataChan          chan any
	rpcQueue             *queue2.SyncQueue
	finish               sync.WaitGroup
	running              *syncx.AtomicBool
	onlineExpired        int64
	clientType           int
	nextNotifyId         int64
	nextPushId           int64
	cb                   *MainAuthWrapperManager
	tmpRpcApiMessageList []*rpcApiMessage
}

func NewMainAuthWrapper(mainAuthId int64, authUserId int64, state int, cb *MainAuthWrapperManager) *MainAuthWrapper {
	mainAuth := &MainAuthWrapper{
		authId:               mainAuthId,
		state:                state,
		AuthUserId:           authUserId,
		mainAuth:             nil,
		mainUpdatesSession:   nil,
		clientType:           0,
		nextPushId:           0,
		nextNotifyId:         math.MaxInt32,
		closeChan:            make(chan struct{}),
		sessionDataChan:      make(chan any, 1024),
		rpcDataChan:          make(chan any, 1024),
		rpcQueue:             queue2.NewSyncQueue(),
		finish:               sync.WaitGroup{},
		running:              syncx.NewAtomicBool(),
		cb:                   cb,
		tmpRpcApiMessageList: make([]*rpcApiMessage, 0),
	}
	mainAuth.mainAuth = newSessionList(mainAuth)

	mainAuth.Start()
	return mainAuth
}

func (m *MainAuthWrapper) changeAuthState(ctx context.Context, state int, stateData interface{}) {
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
		m.AuthUserId = stateData.(int64)
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

		_, _ = m.cb.Service.StatusClient.StatusSetSessionOnline(
			ctx,
			&statuspb.TStatusSetSessionOnline{
				UserId: m.AuthUserId,
				Session: &statuspb.SessionEntry{
					UserId:  m.AuthUserId,
					AuthId:  m.authId,
					Gateway: m.cb.Service.MyServerId,
					Expired: date + 60,
					Client:  "",
				},
			})
		m.onlineExpired = date + 60
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
		logx.WithContext(ctx).Infof("authSessions]]>> offline: %s", m)
		_, _ = m.cb.Service.StatusClient.StatusSetSessionOffline(ctx, &statuspb.TStatusSetSessionOffline{
			UserId: m.AuthUserId,
			AuthId: m.authId,
		})
	}
	m.onlineExpired = 0
}

func (m *MainAuthWrapper) delOnline(ctx context.Context) {
	if m.AuthUserId > 0 {
		logx.Infof("authSessions]]>> delOnline: %s", m)

		_, _ = m.cb.Service.StatusClient.StatusSetSessionOffline(ctx, &statuspb.TStatusSetSessionOffline{
			UserId: m.AuthUserId,
			AuthId: m.authId,
		})
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

func (m *MainAuthWrapper) onUpdateInitConnection(ctx context.Context, clientIp string, initConnection *tproto.InitConnection) {
	if initConnection == nil {
		return
	}

	if m.state < AuthStateUnauthorized {
		m.state = AuthStateUnauthorized
	}
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
	go m.rpcRunLoop()
	go m.runLoop()
}

func (m *MainAuthWrapper) Stop() {
	m.running.Set(false)
	// m.rpcQueue.Close()
}

func (m *MainAuthWrapper) runLoop() {
	defer func() {
		if m.mainUpdatesSession != nil && m.mainUpdatesSession.sessionOnline() {
			m.delOnline(context.Background())
		}
		m.finish.Done()
		m.rpcQueue.Close()
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
					m.onSessionData(ctxData.ctx, &ctxData.sessionData)
				})
			case *sessionHttpDataCtx:
				threading.RunSafe(func() {
					m.onSessionHttpData(ctxData.ctx, &ctxData.sessionHttpData)
				})
			case *syncRpcResultDataCtx:
				threading.RunSafe(func() {
					m.onSyncRpcResultData(ctxData.ctx, &ctxData.syncRpcResultData)
				})
			case *syncDataCtx:
				threading.RunSafe(func() {
					m.onSyncData(ctxData.ctx, &ctxData.syncData)
				})
			case *syncSessionDataCtx:
				threading.RunSafe(func() {
					m.onSyncSessionData(ctxData.ctx, &ctxData.syncSessionData)
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
				panic("receive invalid type msg")
			}
		case rpcMessages := <-m.rpcDataChan:
			threading.RunSafe(func() {
				rpcResult, _ := rpcMessages.(*rpcApiMessage)
				_ = rpcResult
				if sess, ok := rpcResult.sessList.sessions[rpcResult.sessionId]; ok {
					// log.Debugf("onRpcResult result: %s", rpcResult)
					sess.onRpcResult(rpcResult.ctx, rpcResult)
				} else {
					logx.WithContext(rpcResult.ctx).Errorf("onRpcResult - not found rpcSession by sessionId: %d", rpcResult.sessionId)
				}
			})
		case <-ticker.C:
			threading.RunSafe(func() {
				m.onTimer(context.Background())
			})
		}
	}

	logx.Infof("%s -> quit runLoop...", m)
}

func (m *MainAuthWrapper) rpcRunLoop() {
	defer func() {
		close(m.rpcDataChan)
	}()
	for {
		apiRequest := m.rpcQueue.Pop()
		if apiRequest == nil {
			logx.Infof("%s -> quit rpcRunLoop...", m)
			return
		} else {
			threading.RunSafe(func() {
				for _, request := range apiRequest.([]*rpcApiMessage) {
					doRpcRequest(request.ctx,
						m.cb.Service,
						&metadata.RpcMetadata{
							ServerId:    m.cb.Service.MyServerId,
							ClientAddr:  request.clientIp,
							AuthId:      request.sessList.authId,
							SessionId:   request.sessionId,
							ReceiveTime: time.Now().Unix(),
							UserId:      m.AuthUserId,
							ClientMsgId: request.reqMsgId,
							Client:      "",
							LangPack:    "",
						},
						request)
					m.rpcDataChan <- request
				}
			})
		}
	}
}

func (m *MainAuthWrapper) sendToRpcQueue(ctx context.Context, rpcMessage []*rpcApiMessage) {
	m.rpcQueue.Push(rpcMessage)
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

func (m *MainAuthWrapper) SessionClientNew(ctx context.Context, kId int64, gatewayId string, sessionId int64) error {
	cData := &connDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		connData: connData{
			authId:    kId,
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

func (m *MainAuthWrapper) SessionHttpDataArrived(ctx context.Context, kId int64, gatewayId, clientIp string, sessionId int64, buf []byte, resChan chan interface{}) error {
	sData := &sessionHttpDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		sessionHttpData: sessionHttpData{
			authId:     kId,
			gatewayId:  gatewayId,
			clientIp:   clientIp,
			sessionId:  sessionId,
			buf:        buf,
			resChannel: resChan,
		},
	}

	select {
	case m.sessionDataChan <- sData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) SessionClientClosed(ctx context.Context, kType int, kId int64, gatewayId string, sessionId int64) error {
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

func (m *MainAuthWrapper) SyncRpcResultDataArrived(ctx context.Context, kId int64, sessionId, clientMsgId int64, data []byte) error {
	rData := &syncRpcResultDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		syncRpcResultData: syncRpcResultData{
			authType:    0,
			authId:      kId,
			sessionId:   sessionId,
			clientMsgId: clientMsgId,
			data:        data,
		},
	}

	select {
	case m.sessionDataChan <- rData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) SyncSessionDataArrived(ctx context.Context, kId int64, sessionId int64, updates tproto.TObject) error {
	sData := &syncSessionDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		syncSessionData: syncSessionData{
			authType:  0,
			authId:    kId,
			sessionId: sessionId,
			data:      &messageData{obj: updates},
		},
	}

	select {
	case m.sessionDataChan <- sData:
	default:
	}

	return nil
}

func (m *MainAuthWrapper) SyncDataArrived(ctx context.Context, updates tproto.TObject) error {
	sData := &syncDataCtx{
		ctx: contextx.ValueOnlyFrom(ctx),
		syncData: syncData{
			data: &messageData{obj: updates},
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

func (m *MainAuthWrapper) onSessionHttpData(ctx context.Context, sessionMsg *sessionHttpData) {
	sList := m.getSessionList()

	if sList.authId == 0 {
		m.resetAuth(sessionMsg.authId)
	} else if sList.authId != sessionMsg.authId {
		m.resetAuth(sessionMsg.authId)
	}

	TMessage, err := tproto.ConvertBytesToTMessage(sessionMsg.buf)
	if err != nil {
		logx.WithContext(ctx).Errorf("onSessionData - error: {%s}, data: {sessions: %s, gate_id: %d}", err, m, sessionMsg.gatewayId)
		return
	}

	sess, ok := sList.sessions[sessionMsg.sessionId]
	if !ok {
		sess = newSession(sessionMsg.sessionId, sList)
		sList.sessions[sessionMsg.sessionId] = sess
	}

	sess.isHttp = true
	sess.httpQueue.Push(sessionMsg.resChannel)
	sess.onSessionConnNew(ctx, sessionMsg.gatewayId)
	sess.onSessionHttpMessageData(ctx, sessionMsg.gatewayId, sessionMsg.clientIp, TMessage)
}

func (m *MainAuthWrapper) onSessionData(ctx context.Context, sessionMsg *sessionData) {
	sList := m.getSessionList()

	if sList.authId == 0 {
		m.resetAuth(sessionMsg.authId)
	} else if sList.authId != sessionMsg.authId {
		m.resetAuth(sessionMsg.authId)
	}

	TMessage, err := tproto.ConvertBytesToTMessage(sessionMsg.buf)
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
	sess.onSessionMessageData(ctx, sessionMsg.gatewayId, sessionMsg.clientIp, TMessage)
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

func (m *MainAuthWrapper) onSyncRpcResultData(ctx context.Context, syncMsg *syncRpcResultData) {
	logx.WithContext(ctx).Infof("onSyncRpcResultData - receive data: {sess: %s}, data: {auth_id: %d, session_id: %d, client_msg_id: %d}",
		m,
		syncMsg.authId,
		syncMsg.sessionId,
		syncMsg.clientMsgId)

	sess, sessOk := m.mainAuth.sessions[syncMsg.sessionId]

	if !sessOk || sess == nil {
		logx.WithContext(ctx).Errorf("onSyncRpcResultData - not found session by sessionId: %d", syncMsg.sessionId)
		return
	}

	sess.onSyncRpcResultData(ctx, syncMsg.clientMsgId, syncMsg.data)
}

func (m *MainAuthWrapper) onSyncSessionData(ctx context.Context, syncMsg *syncSessionData) {
	logx.WithContext(ctx).Infof("onSyncSessionData - receive data: {sess: %s}",
		m)

	sList := m.getSessionListById(syncMsg.authId)
	if sList == nil {
		logx.WithContext(ctx).Errorf("onSyncRpcResultData - not found sessionList by authId: %d", syncMsg.authId)
		return
	}

	sess, sessOk := sList.sessions[syncMsg.sessionId]
	if sessOk && sess != nil {
		sess.onSyncSessionData(ctx, syncMsg.data.obj)
	}
}

func (m *MainAuthWrapper) onSyncData(ctx context.Context, syncMsg *syncData) {
	logx.WithContext(ctx).Info("authSessions - ", reflect.TypeOf(syncMsg.data.obj))

	if m.mainUpdatesSession != nil {
		m.mainUpdatesSession.onSyncData(ctx, syncMsg.data.obj)
	}
}

func doRpcRequest(ctx context.Context, service *service.Service, md *metadata.RpcMetadata, request *rpcApiMessage) {
	var (
		err       error
		rpcResult tproto.TObject
	)

	rpcResult, err = service.InvokeContext(ctx, md, request.reqMsg)

	reply := &tproto.RpcResult{
		ReqMsgId: request.reqMsgId,
		Result:   nil,
	}

	if err != nil {
		logx.WithContext(ctx).Error(err.Error())
		errResultAny, _ := anypb.New(tproto.NewRpcError(err))
		reply.Result = errResultAny
	} else {
		logx.WithContext(ctx).Infof("invokeRpcRequest - rpc_result: {%s}", reflect.TypeOf(rpcResult))
		rpcResultAny, _ := anypb.New(rpcResult)
		reply.Result = rpcResultAny
	}

	request.rpcResult = reply
}
