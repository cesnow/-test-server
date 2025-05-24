package client

import (
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/config"
)

type Client struct {
	*Rpc
}

type Rpc struct {
	*ShardingSession
}

func New(c config.Config) *Client {
	return &Client{
		&Rpc{
			ShardingSession: NewShardingSession(c),
		},
	}
}
