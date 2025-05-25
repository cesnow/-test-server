package gnet

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/types/known/anypb"
	"hash/crc32"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/gnet/websocket"
	. "kiyudesign.com/cesnow/light-server/app/session/client"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"math/rand"
	"strconv"
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
	logx.Debugf("onNewConn - conn(%s)", c)

	ctx := newConnContext()
	ctx.setClientIp(strings.Split(c.RemoteAddr().String(), ":")[0])
	ctx.wsCodec = new(websocket.Codec)
	ctx.closeDate = time.Now().Unix() + 60
	c.SetContext(ctx)

	fd := int64(c.Fd())
	s.connections.Store(fd, c)
	logx.Infof("onNewConn - conn(%s), connId(%d)", c.RemoteAddr().String(), fd)

	return
}

func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	logx.Debugf("onConnClosed - conn(%s), err: %v", c, err)

	ctx, _ := c.Context().(*connContext)
	if ctx == nil {
		return
	}

	defer func() {
		if ctx.wsCodec != nil {
			ctx.wsCodec.Conn.Release()
		}

		c.SetContext(nil)

		fd := int64(c.Fd())
		s.connections.Delete(fd)

	}()

	if ctx.AuthId() == 0 {
		return
	}

	// kId, sessId, connId, clientIp := ctx.authKey.AuthId(), ctx.sessionId, c.ConnId(), ctx.clientIp
	bDeleted := s.authSessionMgr.RemoveSession(ctx.AuthId(), ctx.sessionId, int64(c.Fd()))
	if !bDeleted {
		return
	}

	_ = s.pool.Submit(func() {
		_ = s.svcCtx.ShardingSession.InvokeByKey(
			strconv.FormatInt(ctx.AuthId(), 10),
			func(client SessionClient) (err error) {
				_, err = client.SessionCloseSession(context.Background(), &sessionpb.TSessionCloseSession{
					Client: &sessionpb.SessionClientEvent{
						ServerId:  s.svcCtx.GatewayId,
						AuthId:    ctx.AuthId(),
						SessionId: ctx.sessionId,
						ClientIp:  ctx.clientIp,
					},
				})
				if err != nil {
					logx.Errorf("client.SessionCloseSession - error: %v", err)
				}
				return
			})
	})

	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := c.Context().(*connContext)
	ctx.closeDate = time.Now().Unix() + 300 + rand.Int63()%10
	return s.onWebsocketData(ctx, c)
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	s.tickNumber = s.tickNumber + 1
	if s.tickNumber%15 == 0 {
		logx.Statf("connection count: %d", s.eng.CountConnections())
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

func (s *Server) onUserReceivedMessage(c gnet.Conn, ctx *connContext, authIdInfo *tproto.AuthIdInfo, msg []byte) error {

	reqAuthId := int64(binary.LittleEndian.Uint64(msg[0:]))
	sessionId := int64(binary.LittleEndian.Uint64(msg[8:]))
	msgId := int64(binary.LittleEndian.Uint64(msg[8+8:]))
	seqNo := int32(binary.LittleEndian.Uint32(msg[8+8+8:]))
	authId := authIdInfo.AuthId

	logx.Infof("onUserReceivedMessage - reqAuthId: %d, authInfo: %d, sessionId: %d, seqNo: %d, msgId: %d, connId: %d",
		reqAuthId, authId, ctx.sessionId, seqNo, msgId, c.Fd())

	if reqAuthId == 0 {
		// hack ping proto, can't send to session
		for _, unknown := range tproto.TryGetUnknownTObject(msg[16:]) {
			if tproto.IsAnyOfType(unknown, &tproto.Ping{}) {
				ping := tproto.Ping{}
				err := unknown.UnmarshalTo(&ping)
				if err != nil {
					logx.Errorf("err unmarshal unknow to ping %s", err)
					continue
				}
				pong, _ := anypb.New(&tproto.Pong{
					MsgId:  msgId,
					PingId: ping.PingId,
				})
				payload := &tproto.TMessage{
					MsgId:  nextMessageId(),
					SeqNo:  seqNo,
					Bytes:  0,
					Object: pong,
				}
				_ = UnThreadSafeWrite(c, serializeToBuffer(0, sessionId, payload))
				break
			}
		}
		// Send to user
		authIdInfoNew, _ := anypb.New(authIdInfo)
		payload := &tproto.TMessage{
			MsgId:  nextMessageId(),
			SeqNo:  seqNo,
			Bytes:  0,
			Object: authIdInfoNew,
		}
		_ = UnThreadSafeWrite(c, serializeToBuffer(authId, sessionId, payload))
		return nil
	}

	var (
		isNew    = ctx.sessionId != sessionId
		clientIp = ctx.clientIp
		connId   = int64(c.Fd())
	)
	if isNew {
		ctx.sessionId = sessionId
	}

	_ = s.pool.Submit(func() {
		_ = s.svcCtx.Client.ShardingSession.InvokeByKey(
			strconv.FormatInt(authId, 10),
			func(client SessionClient) (err error) {
				if isNew {
					logx.Infof("addNewSession - authInfo: %d, sessionId: %d, connId: %d", authId, sessionId, connId)
					if s.authSessionMgr.AddNewSession(authId, sessionId, connId) {
						_, err = client.SessionCreateSession(context.Background(), &sessionpb.TSessionCreateSession{
							Client: &sessionpb.SessionClientEvent{
								ServerId:  s.svcCtx.GatewayId,
								AuthId:    authId,
								SessionId: sessionId,
								ClientIp:  clientIp,
							},
						})
						if err != nil {
							logx.Errorf("client.SessionCreateSession - error: %v", err)
						}
					}
				}
				logx.Infof("SessionSendDataToSession - authInfo: %d, sessionId: %d, connId: %d", authId, sessionId, connId)
				_, err = client.SessionSendDataToSession(context.Background(), &sessionpb.TSessionSendDataToSession{
					Data: &sessionpb.SessionClientData{
						ServerId:  s.svcCtx.GatewayId,
						AuthId:    authId,
						SessionId: sessionId,
						Payload:   msg[16:],
						ClientIp:  clientIp,
					},
				})
				if err != nil {
					logx.Errorf("session.sendDataToSession - error: %v", err)
				}

				return
			})
	})

	return nil
}

func (s *Server) GetConnCounts() int {
	return s.eng.CountConnections()
}

func (s *Server) onReceiveRawMessage(ctx *connContext, c gnet.Conn, authId int64, msg []byte) (action gnet.Action) {

	logx.Debugf("conn(%d-%s) onReceiveRawMessage: %s", c.Fd(), c.RemoteAddr().String(), hex.EncodeToString(msg))

	if authId != 0 {

		authKey := ctx.getAuthIdInfo()
		if authKey == nil {
			data := s.GetAuth(authId)
			if data != nil {
				ctx.putAuthIdInfo(data)
			}
		} else if authKey.GetAuthId() != authId {
			logx.Errorf("conn(%s) getAuthKey - error: invalid key id %d ", c, authId)
			action = gnet.Close
			return
		}

		if authKey != nil {
			err := s.onUserReceivedMessage(c, ctx, authKey, msg)
			if err != nil {
				action = gnet.Close
			}
			return
		}
	}

	msgClone := make([]byte, len(msg))
	copy(msgClone, msg)

	s.asyncRun2(
		int64(c.Fd()),
		msgClone,
		func(msg2 []byte) (any, error) {
			var newAuth *tproto.AuthIdInfo
			err2 := s.svcCtx.ShardingSession.InvokeByKey(
				strconv.FormatInt(authId, 10),
				func(client SessionClient) (err error) {
					newAuth, err = client.SessionQueryAuthId(context.Background(), &sessionpb.TSessionQueryAuthId{})
					return
				})
			if err2 != nil {
				logx.Errorf("conn(%s) sessionQueryAuthId - error: %v", c, err2)
				return nil, err2
			}

			return newAuth, nil
		},
		func(c2 gnet.Conn, msg2 []byte, in any, err error) {
			if err != nil {
				if errors.Is(err, tproto.ErrAuthIdGenerate) {
					// make some err and using serializeToBuffer
					payload := make([]byte, 4)
					binary.LittleEndian.PutUint32(payload, uint32(400))
					_ = UnThreadSafeWrite(c2, payload)
				}
				_ = c2.Close()
			} else {
				authIdInfo := in.(*tproto.AuthIdInfo)
				ctx2 := c2.Context().(*connContext)
				ctx2.putAuthIdInfo(authIdInfo)
				err = s.onUserReceivedMessage(c2, ctx2, authIdInfo, msg2)
				if err != nil {
					logx.Errorf("conn(%s) onUserReceivedMessage - error: %v ", c2, err)
					_ = c2.Close()
				}
			}
		})

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

	err = wsutil.WriteServerBinary(c, data)
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
