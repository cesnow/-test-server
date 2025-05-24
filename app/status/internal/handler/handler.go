package handler

import (
	"context"
	"kiyudesign.com/cesnow/light-server/app/status/internal/application"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type StatusCore struct {
	ctx    context.Context
	svcCtx *application.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *application.ServiceContext) *StatusCore {
	return &StatusCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}
