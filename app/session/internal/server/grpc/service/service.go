package service

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/session/internal/application"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"strconv"
)

type Service struct {
	svcCtx *application.ServiceContext
}

func (s *Service) GetServiceContext() *application.ServiceContext {
	return s.svcCtx
}

func New(ctx *application.ServiceContext) *Service {
	return &Service{
		svcCtx: ctx,
	}
}

func (s *Service) checkShardingV(ctx context.Context, authId int64) (err error) {
	server, ok := s.svcCtx.RpcShardingManager.GetShardingV(strconv.FormatInt(authId, 10))

	if !ok {
		logx.WithContext(ctx).Errorf("not found shardingV by authId: %d", authId)
		err = tproto.ErrInternalServerError
	} else {
		if server != s.svcCtx.Service.MyServerId {
			logx.WithContext(ctx).Errorf("authId(%d), redirect to server: %d", authId, server)
			err = tproto.NewErrRedirectToX(server)
		}
	}

	return
}
