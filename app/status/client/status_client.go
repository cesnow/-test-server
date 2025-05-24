package statusclient

import (
	"context"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	statuspb "kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"

	"github.com/zeromicro/go-zero/zrpc"
)

type StatusClient interface {
	StatusSetSessionOnline(ctx context.Context, in *statuspb.TStatusSetSessionOnline) (*tproto.Bool, error)
	StatusSetSessionOffline(ctx context.Context, in *statuspb.TStatusSetSessionOffline) (*tproto.Bool, error)
	StatusGetUserOnlineSessions(ctx context.Context, in *statuspb.TStatusGetUserOnlineSessions) (*statuspb.UserSessionEntryList, error)
}

type defaultStatusClient struct {
	cli zrpc.Client
}

func NewStatusClient(cli zrpc.Client) StatusClient {
	return &defaultStatusClient{
		cli: cli,
	}
}

func (m *defaultStatusClient) StatusSetSessionOnline(ctx context.Context, in *statuspb.TStatusSetSessionOnline) (*tproto.Bool, error) {
	client := statuspb.NewRpcStatusClient(m.cli.Conn())
	return client.StatusSetSessionOnline(ctx, in)
}

func (m *defaultStatusClient) StatusSetSessionOffline(ctx context.Context, in *statuspb.TStatusSetSessionOffline) (*tproto.Bool, error) {
	client := statuspb.NewRpcStatusClient(m.cli.Conn())
	return client.StatusSetSessionOffline(ctx, in)
}

func (m *defaultStatusClient) StatusGetUserOnlineSessions(ctx context.Context, in *statuspb.TStatusGetUserOnlineSessions) (*statuspb.UserSessionEntryList, error) {
	client := statuspb.NewRpcStatusClient(m.cli.Conn())
	return client.StatusGetUserOnlineSessions(ctx, in)
}
