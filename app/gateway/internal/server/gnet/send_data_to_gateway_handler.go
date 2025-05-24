package gnet

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"

	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Server) GatewaySendDataToGateway(ctx context.Context, in *gatewaypb.TGatewaySendDataToGateway) (reply *tproto.Bool, err error) {

	logx.WithContext(ctx).Infof("ReceiveData - request: {kId: %d, sessionId: %d, payloadLen: %d}", in.AuthId, in.SessionId, len(in.Payload))

	authId, connIdList := s.authSessionMgr.FoundSessionConnId(in.AuthId, in.SessionId)
	if len(connIdList) == 0 {
		logx.WithContext(ctx).Errorf("ReceiveData - not found connId - AuthId: %d, sessionId: %d", in.AuthId, in.SessionId)
		return tproto.BoolFalse, nil
	} else {
		logx.WithContext(ctx).Debugf("found: {k: %v, idList: %v}", in.AuthId, connIdList)
	}

	ctx = contextx.ValueOnlyFrom(ctx)
	msg := serializeUserToBuffer(in.AuthId, in.SessionId, in.Payload)

	_ = s.pool.Submit(func() {
		for _, connId := range connIdList {
			s.Trigger(connId, func(c gnet.Conn) {
				connCtx, _ := c.Context().(*connContext)
				if connCtx == nil {
					logx.WithContext(ctx).Errorf("invalid state - conn(%s) Context() is nil", c)
					return
				}

				if in.AuthId != connCtx.AuthId() {
					logx.WithContext(ctx).Errorf("invalid state - conn(%s) c.keyId(%d) != in.keyId(%d) is nil", authId, in.AuthId)
					return
				}

				err2 := UnThreadSafeWrite(c, msg)
				if err2 != nil {
					logx.WithContext(ctx).Errorf("sendToClient error: %v", err2)
				} else {
					logx.WithContext(ctx).Debugf("sendToConn: %v", connId)
				}
			})
		}
	})

	return tproto.BoolTrue, nil
}
