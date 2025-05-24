package service

import (
	"kiyudesign.com/cesnow/light-server/app/status/internal/application"
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
