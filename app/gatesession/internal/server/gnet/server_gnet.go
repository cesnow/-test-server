package gnet

import (
	"context"
	"encoding/hex"
	"github.com/panjf2000/gnet/v2"
	"github.com/segmentio/ksuid"
	"github.com/vmihailenco/msgpack/v5"
	"hash/crc32"
	"hash/fnv"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/handler"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/sse"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/websocket"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/service/authsession"
	"kiyudesign.com/cesnow/light-server/pkg/transport"
	"math/rand"
	"strings"
	"time"

	"github.com/gobwas/ws/wsutil"
	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Server) asyncRun(connId int64, exec func() error, returnCB func(c gnet.Conn)) {
	_ = s.pool.Submit(func() {
		if err := exec(); err == nil {
			s.Trigger(connId, func(c gnet.Conn) {
				returnCB(c)
			})
		}
	})
}

func (s *Server) asyncRun2(
	connId int64,
	msg []byte,
	exec func(msg []byte) (any, error),
	returnCB func(c gnet.Conn, msg []byte, in any, err error)) {
	_ = s.pool.Submit(func() {
		r, err := exec(msg)
		s.Trigger(connId, func(c gnet.Conn) {
			returnCB(c, msg, r, err)
		})
	})
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	logx.Infof("gnetway server is listening")
	s.eng = eng
	return gnet.None
}

func (s *Server) OnShutdown(eng gnet.Engine) {
	_ = eng
	logx.Infof("gnetway server shutdown")
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	logx.Debugf("onNewConn - conn(fd:%d, addr:%s)", c.Fd(), c.RemoteAddr().String())

	ctx := newConnContext()
	ctx.setClientIp(strings.Split(c.RemoteAddr().String(), ":")[0])

	ctx.sseCodec = new(sse.Codec)
	ctx.wsCodec = new(websocket.Codec)

	ctx.closeDate = time.Now().Unix() + 120
	c.SetContext(ctx)

	fd := int64(c.Fd())
	s.connections.Store(fd, c)
	logx.Infof("onNewConn - conn(%s), connId(%d)", c.RemoteAddr().String(), fd)

	return
}

func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	logx.Debugf("onConnClosed - conn(fd: %d, addr: %s), err: %v", c, c.RemoteAddr().String(), err)

	ctx, _ := c.Context().(*connContext)
	if ctx == nil {
		return
	}

	defer func() {
		if ctx.connType == ConnectionTypeWebSocket && ctx.wsCodec != nil {
			ctx.wsCodec.Conn.Release()
		}

		if ctx.connType == ConnectionTypeSSE && ctx.sseCodec != nil {
			// pass
		}

		c.SetContext(nil)

		fd := int64(c.Fd())
		s.connections.Delete(fd)

	}()

	if ctx.AuthId() == 0 {
		return
	}

	bDeleted := s.authSessionMgr.RemoveSession(ctx.AuthId(), ctx.sessionId, int64(c.Fd()))
	if !bDeleted {
		return
	}

	mainAuth := s.svcCtx.MainAuthMgr.GetMainAuthWrapper(ctx.AuthId())
	if mainAuth == nil {
		logx.Errorf("session.closeSession - not found sessList by keyId: %d", ctx.AuthId())
	} else {
		_ = mainAuth.CloseSession(context.Background(), ctx.AuthId(), ctx.clientIp, ctx.sessionId)
	}

	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := c.Context().(*connContext)
	ctx.closeDate = time.Now().Unix() + 300 + rand.Int63()%10

	return s.onSSEData(ctx, c)
	//switch ctx.connType {
	//case ConnectionTypeSSE:
	//	return s.onSSEData(ctx, c)
	//case ConnectionTypeWebSocket:
	//	return s.onSSEData(ctx, c)
	//default:
	//	return gnet.Close
	//}
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	s.tickNumber = s.tickNumber + 1
	if s.tickNumber%15 == 0 {
		logx.Statf("connection count: %d", s.eng.CountConnections())
	}

	if s.tickNumber%30 == 0 {
		s.sendSSEHeartbeat()
	}

	delay = time.Second * 1
	now := time.Now().Unix()

	s.Iterate(func(c gnet.Conn) {
		ctx, _ := c.Context().(*connContext)
		if ctx == nil {
			return
		}
		if now >= ctx.closeDate {
			logx.Errorf("close conn(%d-%s) by timeout", c.Fd(), c.RemoteAddr().String())
			_ = c.Close()
		}
	})
	return
}

func (s *Server) GetConnCounts() int {
	return s.eng.CountConnections()
}

func (s *Server) getOrFetchMainAuthWrapper(mainAuthId int64) (*authsession.MainAuthWrapper, error) {
	mainAuth := s.svcCtx.MainAuthMgr.GetMainAuthWrapper(mainAuthId)
	if mainAuth != nil {
		return mainAuth, nil
	}

	mainAuth = s.svcCtx.MainAuthMgr.AllocMainAuthWrapper(
		mainAuthId,
		func(authId int64) *authsession.MainAuthWrapper {
			eventHandler := handler.New(context.Background(), s.svcCtx)
			return authsession.NewMainAuthWrapper(
				mainAuthId,
				0,
				authsession.AuthStateNew,
				s.svcCtx.MainAuthMgr,
				s.SendDataToClient,
				eventHandler.EventCall)
		},
	)

	return mainAuth, nil
}

func (s *Server) onReceiveRawMessage(ctx *connContext, c gnet.Conn, msg []byte) (action gnet.Action) {

	logx.Debugf("conn(%d-%s) onReceiveRawMessage: [%d] %s", c.Fd(), c.RemoteAddr().String(), len(msg), hex.EncodeToString(msg))

	ctxAuthId := ctx.AuthId()
	if ctxAuthId == 0 {
		nFnv := fnv.New64()
		_, err := nFnv.Write(ksuid.New().Bytes())
		authId := int64(nFnv.Sum64())
		if err != nil {
			logx.Errorf("conn(%s) getAuthKey - error: %v ", c, err)
			action = gnet.Close
			return
		}
		//data := s.GetAuth(authId)
		//if data != nil {
		//	ctx.putAuthId(data)
		//}
		logx.Infof("conn(%d-%s) generate authId: %d", c.Fd(), c.RemoteAddr().String(), authId)
		ctx.putAuthId(authId)
	}

	// process the inner message
	var tMsg *transport.TMsgRawData
	err := msgpack.Unmarshal(msg, &tMsg)
	if err != nil {
		logx.Errorf("onSessionData - error: {%s}", err)
		return gnet.None
	}

	switch tMsg.Event {
	case "ping":
		pong := transport.Pong{PingId: 10}
		pongBuf, _ := msgpack.Marshal(&pong)
		tMsg.Event = "pong"
		tMsg.Body = pongBuf
		tMsg.ReqMsgId = tMsg.MsgId
		tMsgBuf, _ := msgpack.Marshal(tMsg)
		_ = UnThreadSafeWrite(c, tMsgBuf)
		return gnet.None
	default:

	}

	var (
		isNew    = ctx.sessionId != 1
		clientIp = ctx.clientIp
		connId   = int64(c.Fd())
		authId   = ctx.AuthId()
	)
	if isNew {
		ctx.sessionId = 1
	}

	_ = s.pool.Submit(func() {
		mainAuth, _ := s.getOrFetchMainAuthWrapper(authId)
		if isNew {
			logx.Infof("addNewSession - ctxAuthId: %d, sessionId: %d, connId: %d", authId, 1, connId)
			if s.authSessionMgr.AddNewSession(authId, 1, connId) {
				_ = mainAuth.NewSession(context.Background(), authId, s.svcCtx.GatewayId, 1)
			}
		}
		_ = mainAuth.SessionDataArrived(
			context.Background(),
			authId,
			s.svcCtx.GatewayId,
			clientIp,
			1,
			msg)
	})

	//go func() {
	//	ticker := time.NewTicker(5 * time.Second)
	//	defer ticker.Stop()
	//	mainAuth, _ := s.getOrFetchMainAuthWrapper(authId)
	//
	//	for range ticker.C {
	//		_ = mainAuth.SyncDataArrived(context.Background(), []byte{0, 0, 0, 0, 0, 0, 0, 0})
	//	}
	//}()

	return gnet.None
}

func UnThreadSafeWrite(c gnet.Conn, msg any) error {
	ctx := c.Context().(*connContext)

	if ctx.codec == nil {
		logx.Errorf("conn(%s) c.Write(data) - error: ctx.codec == nil ", c)
		return nil
	}

	if msg == nil {
		logx.Errorf("conn(%s) c.Write(data) - error: msg == ni ", c)
		return nil
	}

	data, err := ctx.codec.Encode(c, msg)
	if err != nil {
		logx.Errorf("conn(%s) ctx.codec.Encode(c, msg) - error: %v ", c, err)
		return err
	}

	if ctx.connType == ConnectionTypeWebSocket {
		err = wsutil.WriteServerBinary(c, data)
	} else if ctx.connType == ConnectionTypeSSE {
		var sseData []byte
		sseData, err = ctx.sseCodec.Encode("data", data)
		_, err = c.Write(sseData)
	}

	if err != nil {
		logx.Errorf("conn[%v] [err=%v]", c.RemoteAddr().String(), err.Error())
		return err
	}

	return nil
}

func (s *Server) GetConnIdx(c gnet.Conn) uint32 {
	h := crc32.ChecksumIEEE([]byte(c.RemoteAddr().String()))
	idx := int(h % uint32(s.numEventLoop))
	return uint32(idx)
}

func (s *Server) Trigger(fd int64, cb func(gnet.Conn)) {
	if conn, ok := s.connections.Load(fd); ok {
		c := conn.(gnet.Conn)
		cb(c)
	} else {
		logx.Errorf("connection not found: fd=%d\n", fd)
	}
}

func (s *Server) Iterate(cb func(c gnet.Conn)) {
	s.connections.Range(func(key, value any) bool {
		cb(value.(gnet.Conn))
		return true
	})
}
