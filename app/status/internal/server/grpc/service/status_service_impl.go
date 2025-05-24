package service

import (
	"context"
	"kiyudesign.com/cesnow/light-server/app/status/internal/handler"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
)

func (s *Service) StatusSetSessionOnline(ctx context.Context, request *statuspb.TStatusSetSessionOnline) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("status.setSessionOnline - metadata: %s, request: %s", c.MD, request)

	r, err := c.StatusSetSessionOnline(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("status.setSessionOnline - reply: %s", r)
	return r, err
}

func (s *Service) StatusSetSessionOffline(ctx context.Context, request *statuspb.TStatusSetSessionOffline) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("status.setSessionOffline - metadata: %s, request: %s", c.MD, request)

	r, err := c.StatusSetSessionOffline(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("status.setSessionOffline - reply: %s", r)
	return r, err
}

func (s *Service) StatusGetUserOnlineSessions(ctx context.Context, request *statuspb.TStatusGetUserOnlineSessions) (*statuspb.UserSessionEntryList, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("status.getUserOnlineSessions - metadata: %s, request: %s", c.MD, request)

	r, err := c.StatusGetUserOnlineSessions(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("status.getUserOnlineSessions - reply: %s", r)
	return r, err
}
