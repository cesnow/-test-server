package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
)

func (c *SyncHandler) SyncPushRpcResult(in *syncpb.TSyncPushRpcResult) (*tproto.Void, error) {
	_ = c.svcCtx.Service.PushRpcResultToSession(c.ctx, in.ServerId, &sessionpb.TSessionPushRpcResultData{
		AuthId:         in.AuthId,
		SessionId:      in.SessionId,
		ClientReqMsgId: in.ClientReqMsgId,
		RpcResultData:  in.RpcResult,
	})

	return &tproto.Void{}, nil
}
