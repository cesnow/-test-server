package service

import (
	"context"
	"kiyudesign.com/cesnow/light-server/app/session/internal/handler"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (s *Service) SessionQueryAuthId(ctx context.Context, request *sessionpb.TSessionQueryAuthId) (*tproto.AuthIdInfo, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.QueryAuthId - metadata: %s, request: %s", c.MD, request)

	r, err := c.SessionQueryAuthId(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.createSession - reply: %s", r)
	return r, err
}

func (s *Service) SessionCreateSession(ctx context.Context, request *sessionpb.TSessionCreateSession) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.createSession - metadata: %s, request: %s", c.MD, request)

	if err := s.checkShardingV(ctx, request.GetClient().GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionCreateSession(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.createSession - reply: %s", r)
	return r, err
}

func (s *Service) SessionSendDataToSession(ctx context.Context, request *sessionpb.TSessionSendDataToSession) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.sendDataToSession - request: {server_id: %s, auth_key_id: %d, session_id: %d, client_ip: %s, payload: %d}",
		request.GetData().GetServerId(),
		request.GetData().GetAuthId(),
		request.GetData().GetSessionId(),
		request.GetData().GetClientIp(),
		len(request.GetData().GetPayload()))

	if err := s.checkShardingV(ctx, request.GetData().GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionSendDataToSession(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.sendDataToSession - reply: %s", r)
	return r, err
}

func (s *Service) SessionSendHttpDataToSession(ctx context.Context, request *sessionpb.TSessionSendHttpDataToSession) (*sessionpb.HttpSessionData, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.sendHttpDataToSession - request: {server_id: %s, conn_type: %d, auth_id: %d, session_id: %d, client_ip: %s, payload: %d}",
		request.GetClient().GetServerId(),
		request.GetClient().GetConnType(),
		request.GetClient().GetAuthId(),
		request.GetClient().GetSessionId(),
		request.GetClient().GetClientIp(),
		len(request.GetClient().GetPayload()))

	if err := s.checkShardingV(ctx, request.GetClient().GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionSendHttpDataToSession(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.sendHttpDataToSession - reply: %s", r)
	return r, err
}

func (s *Service) SessionCloseSession(ctx context.Context, request *sessionpb.TSessionCloseSession) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.closeSession - metadata: %s, request: %s", c.MD, request)

	if err := s.checkShardingV(ctx, request.GetClient().GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionCloseSession(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.closeSession - reply: %s", r)
	return r, err
}

func (s *Service) SessionPushUpdatesData(ctx context.Context, request *sessionpb.TSessionPushUpdatesData) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.pushUpdatesData - metadata: %s, request: %s", c.MD, request)

	if err := s.checkShardingV(ctx, request.GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionPushUpdatesData(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.pushUpdatesData - reply: %s", r)
	return r, err
}

func (s *Service) SessionPushSessionUpdatesData(ctx context.Context, request *sessionpb.TSessionPushSessionUpdatesData) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.pushSessionUpdatesData - metadata: %s, request: %s", c.MD, request)

	if err := s.checkShardingV(ctx, request.GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionPushSessionUpdatesData(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.pushSessionUpdatesData - reply: %s", r)
	return r, err
}

func (s *Service) SessionPushRpcResultData(ctx context.Context, request *sessionpb.TSessionPushRpcResultData) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("session.pushRpcResultData - metadata: %s, request: %s", c.MD, request)

	if err := s.checkShardingV(ctx, request.GetAuthId()); err != nil {
		return nil, err
	}

	r, err := c.SessionPushRpcResultData(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("session.pushRpcResultData - reply: %s", r)
	return r, err
}
