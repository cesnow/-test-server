package gnet

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/contextx"
	"kiyudesign.com/cesnow/light-server/pkg/transport"

	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Server) SendDataToClient(ctx context.Context, gatewayId string, authId int64, sessionId int64, data *transport.TMsgRawData) (bool, error) {

	logx.Infof("sendToClient - gatewayId: %s, authId: %d, sessionId: %d, dataLen: %v", gatewayId, authId, sessionId, len(data.Body))

	resultData, _ := msgpack.Marshal(data)

	_, connIdList := s.authSessionMgr.FoundSessionConnId(authId, sessionId)
	if len(connIdList) == 0 {
		logx.WithContext(ctx).Errorf("ReceiveData - not found connId - AuthId: %d, sessionId: %d", authId, sessionId)
		return false, nil
	} else {
		logx.WithContext(ctx).Debugf("found: {k: %v, idList: %v}", authId, connIdList)
	}

	ctx = contextx.ValueOnlyFrom(ctx)

	_ = s.pool.Submit(func() {
		for _, connId := range connIdList {
			s.Trigger(connId, func(c gnet.Conn) {
				connCtx, _ := c.Context().(*connContext)
				if connCtx == nil {
					logx.WithContext(ctx).Errorf("invalid state - conn(%s) Context() is nil", c)
					return
				}

				if authId != connCtx.AuthId() {
					logx.WithContext(ctx).Errorf("invalid state - conn(%d) c.keyId(%d) != authId(%d) is nil", connId, connCtx.AuthId(), authId)
					return
				}

				err2 := UnThreadSafeWrite(c, resultData)
				if err2 != nil {
					logx.WithContext(ctx).Errorf("sendToClient error: %v", err2)
				} else {
					logx.WithContext(ctx).Debugf("sendToConn: %v", connId)
				}
			})
		}
	})

	return true, nil
}
