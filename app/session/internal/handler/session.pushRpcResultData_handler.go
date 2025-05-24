package handler

import (
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (c *SessionHandler) SessionPushRpcResultData(in *sessionpb.TSessionPushRpcResultData) (*tproto.Bool, error) {
	mainAuth := c.svcCtx.MainAuthMgr.GetMainAuthWrapper(in.AuthId)
	if mainAuth == nil {
		err := fmt.Errorf("not found authId(%s)", in)
		c.Logger.Errorf("session.pushRpcResultData - %v", err)
		return nil, err
	}
	_ = mainAuth.SyncRpcResultDataArrived(c.ctx, in.AuthId, in.SessionId, in.ClientReqMsgId, in.RpcResultData)

	return tproto.BoolTrue, nil
}
