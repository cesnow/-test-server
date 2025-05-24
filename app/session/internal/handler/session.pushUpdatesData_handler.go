package handler

import (
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (c *SessionHandler) SessionPushUpdatesData(in *sessionpb.TSessionPushUpdatesData) (*tproto.Bool, error) {
	mainAuth := c.svcCtx.MainAuthMgr.GetMainAuthWrapper(in.AuthId)
	if mainAuth == nil {
		err := fmt.Errorf("not found authId(%d)", in.AuthId)
		c.Logger.Errorf("session.pushUpdatesData - %v", err)
		return nil, err
	}
	mainAuth.SyncDataArrived(c.ctx, in.Updates)

	return tproto.BoolTrue, nil
}
