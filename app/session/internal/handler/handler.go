package handler

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/session/internal/application"
	"kiyudesign.com/cesnow/light-server/app/session/internal/logic"
	metadata2 "kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"
)

type SessionHandler struct {
	ctx    context.Context
	svcCtx *application.ServiceContext
	logx.Logger
	MD *metadata2.RpcMetadata
}

func New(ctx context.Context, svcCtx *application.ServiceContext) *SessionHandler {
	return &SessionHandler{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata2.RpcMetadataFromIncoming(ctx),
	}
}

func (c *SessionHandler) getOrFetchMainAuthWrapper(mainAuthId int64) (*logic.MainAuthWrapper, error) {
	mainAuth := c.svcCtx.MainAuthMgr.GetMainAuthWrapper(mainAuthId)
	if mainAuth != nil {
		return mainAuth, nil
	}

	mainAuth = c.svcCtx.MainAuthMgr.AllocMainAuthWrapper(
		mainAuthId,
		func(authId int64) *logic.MainAuthWrapper {
			return logic.NewMainAuthWrapper(
				mainAuthId,
				0,
				logic.AuthStateNew,
				c.svcCtx.MainAuthMgr)
		},
	)

	return mainAuth, nil
}
