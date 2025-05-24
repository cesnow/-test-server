package service

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"
)

func (s *Service) GatewaySendDataToGateway(ctx context.Context, request *gatewaypb.TGatewaySendDataToGateway) (reply *tproto.Bool, err error) {
	logx.WithContext(ctx).Debugf("gateway.sendDataToGateway - request: {auth_key_id:%d, session_id:long:%d, payload: %d}",
		request.AuthId,
		request.SessionId,
		len(request.Payload))

	r, err := s.RpcGatewayServer.GatewaySendDataToGateway(ctx, request)
	if err != nil {
		return nil, err
	}

	logx.WithContext(ctx).Debugf("gateway.sendDataToGateway - reply: %s", r)
	return r, err
}
