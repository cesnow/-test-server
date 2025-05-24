package configurationclient

import (
	"context"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"

	"github.com/zeromicro/go-zero/zrpc"
)

type ConfigurationClient interface {
	HelpGetCountriesList(ctx context.Context, in *configurationpb.THelpGetCountriesList) (*configurationpb.Help_CountriesList, error)
}

type defaultConfigurationClient struct {
	cli zrpc.Client
}

func NewConfigurationClient(cli zrpc.Client) ConfigurationClient {
	return &defaultConfigurationClient{
		cli: cli,
	}
}

func (m *defaultConfigurationClient) HelpGetCountriesList(ctx context.Context, in *configurationpb.THelpGetCountriesList) (*configurationpb.Help_CountriesList, error) {
	client := configurationpb.NewRpcConfigurationClient(m.cli.Conn())
	return client.HelpGetCountriesList(ctx, in)
}
