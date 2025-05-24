package http

import (
	"context"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Server) GatewaySendDataToGateway(ctx context.Context, in *gatewaypb.TGatewaySendDataToGateway) (reply *tproto.Bool, err error) {
	logx.WithContext(ctx).Infof("ReceiveData - request: {kId: %d, sessionId: %d, payloadLen: %d}", in.AuthId, in.SessionId, len(in.Payload))

	userID := fmt.Sprintf("%d", in.AuthId)

	s.sseMutex.RLock()
	clientChan, ok := s.sseClients[userID]
	s.sseMutex.RUnlock()

	if !ok {
		logx.WithContext(ctx).Errorf("No SSE client found for userID: %s", userID)
		return tproto.BoolFalse, nil
	}

	select {
	case clientChan <- string(in.Payload):
		logx.WithContext(ctx).Infof("Data sent to userID: %s", userID)
	case <-time.After(2 * time.Second):
		logx.WithContext(ctx).Errorf("Timeout while sending data to userID: %s", userID)
		return tproto.BoolFalse, nil
	}

	return tproto.BoolTrue, nil
}
