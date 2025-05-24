package handler

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"kiyudesign.com/cesnow/light-server/app/sync/internal/application"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
)

type SyncType int

type SyncHandler struct {
	ctx    context.Context
	svcCtx *application.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *application.ServiceContext) *SyncHandler {
	return &SyncHandler{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}

func (c *SyncHandler) pushUpdatesToSession(userId int64, hasServerId *wrapperspb.StringValue, authId, sessionId *wrapperspb.Int64Value, pushData *anypb.Any) {

	var (
		serverIdKeyIdList = make(map[string][]int64)
	)

	statusList, _ := c.svcCtx.Service.StatusClient.StatusGetUserOnlineSessions(c.ctx, &statuspb.TStatusGetUserOnlineSessions{
		UserId: userId,
	})
	c.Logger.Debugf("statusList - #%v", statusList)
	for _, sess := range statusList.GetUserSessions() {
		if keyIdList, ok := serverIdKeyIdList[sess.Gateway]; ok {
			keyIdList = append(keyIdList, sess.AuthId)
			serverIdKeyIdList[sess.Gateway] = keyIdList
		} else {
			serverIdKeyIdList[sess.Gateway] = []int64{sess.AuthId}
		}
	}

	c.Logger.Debugf("serverIdKeyIdList - #%v", serverIdKeyIdList)
	for serverId, keyIdList := range serverIdKeyIdList {
		for _, keyId := range keyIdList {
			// log.Debugf("serverIdKeyIdList - #%v", serverIdKeyIdList)
			_ = c.svcCtx.Service.PushUpdatesToSession(
				c.ctx,
				serverId,
				&sessionpb.TSessionPushUpdatesData{
					AuthId:  keyId,
					Updates: pushData,
				})
		}
	}

}
