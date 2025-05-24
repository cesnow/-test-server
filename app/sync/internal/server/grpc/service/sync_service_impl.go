package service

import (
	"context"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/handler"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
)

func (s *Service) SyncPushUpdates(ctx context.Context, request *syncpb.TSyncPushUpdates) (*tproto.Void, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("sync.pushUpdates - metadata: %s, request: %s", c.MD, request)

	r, err := c.SyncPushUpdates(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("sync.pushUpdates - reply: %s", r)
	return r, err
}

func (s *Service) SyncPushRpcResult(ctx context.Context, request *syncpb.TSyncPushRpcResult) (*tproto.Void, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("sync.pushRpcResult - metadata: %s, request: %s", c.MD, request)

	r, err := c.SyncPushRpcResult(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("sync.pushRpcResult - reply: %s", r)
	return r, err
}
