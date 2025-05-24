package handler

import (
	"context"
	"google.golang.org/protobuf/types/known/anypb"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
)

func (c *ConfigurationHandler) HelpTryNotifyUser(in *configurationpb.THelpTryNotifyUser) (*tproto.Bool, error) {

	testMsg := &configurationpb.THelpTestReceiveMsg{
		FromUserId: c.MD.UserId,
		Message:    in.Message,
	}
	anyMsg, _ := anypb.New(testMsg)

	_, _ = c.svcCtx.Service.SyncPushUpdates(context.Background(), &syncpb.TSyncPushUpdates{
		UserId:  in.UserId,
		Updates: anyMsg,
	})

	return tproto.BoolTrue, nil
}
