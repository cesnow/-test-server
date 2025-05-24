package service

import (
	"context"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/handler"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
)

func (s *Service) HelpGetCountriesList(ctx context.Context, request *configurationpb.THelpGetCountriesList) (*configurationpb.Help_CountriesList, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("help.getCountriesList - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.HelpGetCountriesList(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("help.getCountriesList - reply: {%s}", r)
	return r, err
}

func (s *Service) HelpTryNotifyUser(ctx context.Context, request *configurationpb.THelpTryNotifyUser) (*tproto.Bool, error) {
	c := handler.New(ctx, s.svcCtx)
	c.Logger.Debugf("help.getCountriesList - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.HelpTryNotifyUser(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("help.tryNotifyUser - reply: {%s}", r)
	return r, err
}
