package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/application"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/handler"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func New(svcCtx *application.ServiceContext, conf kafka.KafkaConsumerConf) *kafka.ConsumerGroup {
	s := kafka.MustKafkaConsumer(&conf)
	s.RegisterHandlers(
		conf.Topics[0],
		func(ctx context.Context, method, key string, value []byte) {
			logx.WithContext(ctx).Debugf("method: %s, key: %s, value: %s", key, value)

			switch protoreflect.FullName(method) {
			case proto.MessageName((*syncpb.TSyncPushUpdates)(nil)):
				threading.RunSafe(func() {
					c := handler.New(ctx, svcCtx)

					r := new(syncpb.TSyncPushUpdates)
					if err := json.Unmarshal(value, r); err != nil {
						c.Logger.Error(err.Error())
						return
					}
					c.Logger.Debugf("sync.pushUpdates - request: %s", r)

					_, _ = c.SyncPushUpdates(r)
				})
			case proto.MessageName((*syncpb.TSyncPushRpcResult)(nil)):
				threading.RunSafe(func() {
					c := handler.New(ctx, svcCtx)

					r := new(syncpb.TSyncPushRpcResult)
					if err := json.Unmarshal(value, r); err != nil {
						c.Logger.Error(err.Error())
						return
					}
					c.Logger.Debugf("sync.pushRpcResult - request: %s", r)

					_, _ = c.SyncPushRpcResult(r)
				})
			default:
				err := fmt.Errorf("invalid key: %s", key)
				logx.Error(err.Error())
			}
		})
	return s
}
