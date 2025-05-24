package syncclient

import (
	"context"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"

	"github.com/zeromicro/go-zero/zrpc"
)

var _ *tproto.Bool

type SyncClient interface {
	SyncPushUpdates(ctx context.Context, in *syncpb.TSyncPushUpdates) (*tproto.Void, error)
	SyncPushRpcResult(ctx context.Context, in *syncpb.TSyncPushRpcResult) (*tproto.Void, error)
}

type defaultSyncClient struct {
	cli zrpc.Client
}

func NewSyncClient(cli zrpc.Client) SyncClient {
	return &defaultSyncClient{
		cli: cli,
	}
}

func (m *defaultSyncClient) SyncPushUpdates(ctx context.Context, in *syncpb.TSyncPushUpdates) (*tproto.Void, error) {
	client := syncpb.NewRpcSyncClient(m.cli.Conn())
	return client.SyncPushUpdates(ctx, in)
}

func (m *defaultSyncClient) SyncPushRpcResult(ctx context.Context, in *syncpb.TSyncPushRpcResult) (*tproto.Void, error) {
	client := syncpb.NewRpcSyncClient(m.cli.Conn())
	return client.SyncPushRpcResult(ctx, in)
}
