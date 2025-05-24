package service

import (
	"github.com/valkey-io/valkey-go"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/status/internal/config"
)

type Service struct {
	KV valkey.Client
}

func New(c config.Config) *Service {

	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{c.Status[0].Host},
		ShuffleInit: true,
	})
	if err != nil {
		logx.Errorf("failed to create valkey client: %v", err)
		panic(err)
	}

	d := &Service{
		KV: client,
	}

	return d
}
