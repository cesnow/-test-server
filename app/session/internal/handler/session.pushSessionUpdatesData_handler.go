package handler

import (
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (c *SessionHandler) SessionPushSessionUpdatesData(in *sessionpb.TSessionPushSessionUpdatesData) (*tproto.Bool, error) {
	mainAuth := c.svcCtx.MainAuthMgr.GetMainAuthWrapper(in.AuthId)
	if mainAuth == nil {
		err := fmt.Errorf("not found authId(%d)", in.AuthId)
		c.Logger.Errorf("session.pushSessionUpdatesData - %v", err)
		return nil, err
	}

	_ = mainAuth.SyncSessionDataArrived(c.ctx, in.AuthId, in.SessionId, in.Updates)

	return tproto.BoolTrue, nil
}
