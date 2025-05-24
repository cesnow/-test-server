package service

import (
	"kiyudesign.com/cesnow/light-server/app/sync/internal/application"
)

type Service struct {
	svcCtx *application.ServiceContext
}

func New(ctx *application.ServiceContext) *Service {
	return &Service{
		svcCtx: ctx,
	}
}
