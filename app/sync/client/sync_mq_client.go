package syncclient

import (
	"context"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
	"strconv"

	"github.com/zeromicro/go-zero/core/jsonx"
	"google.golang.org/protobuf/proto"
)

type defaultSyncMqClient struct {
	cli *kafka.Producer
}

func NewSyncMqClient(cli *kafka.Producer) SyncClient {
	return &defaultSyncMqClient{
		cli: cli,
	}
}

func (m *defaultSyncMqClient) sendMessage(ctx context.Context, method, k string, in interface{}) (*tproto.Void, error) {
	var (
		b   []byte
		err error
	)

	b, err = jsonx.Marshal(in)
	if err != nil {
		return nil, err
	}

	_, _, err = m.cli.SendMessageV2(ctx, method, k, b)
	if err != nil {
		return nil, err
	}

	return &tproto.Void{}, nil
}

func (m *defaultSyncMqClient) SyncPushUpdates(ctx context.Context, in *syncpb.TSyncPushUpdates) (*tproto.Void, error) {
	return m.sendMessage(
		ctx,
		string(proto.MessageName(in)),
		strconv.FormatInt(in.GetUserId(), 10),
		in)
}

func (m *defaultSyncMqClient) SyncPushRpcResult(ctx context.Context, in *syncpb.TSyncPushRpcResult) (*tproto.Void, error) {
	return m.sendMessage(
		ctx,
		string(proto.MessageName(in)),
		strconv.FormatInt(in.GetAuthId(), 10),
		in)
}
