package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (c *SessionHandler) SessionCloseSession(in *sessionpb.TSessionCloseSession) (*tproto.Bool, error) {
	var (
		cli = in.GetClient()
	)

	if cli == nil {
		err := tproto.ErrInputRequestInvalid
		c.Logger.Errorf("session.closeSession - error: %v", err)
		return nil, err
	}

	mainAuth := c.svcCtx.MainAuthMgr.GetMainAuthWrapper(cli.AuthId)
	if mainAuth == nil {
		c.Logger.Errorf("session.closeSession - not found sessList by keyId: %s", cli)
	} else {
		_ = mainAuth.SessionClientClosed(c.ctx, int(0), cli.AuthId, cli.ServerId, cli.SessionId)
	}

	return tproto.BoolTrue, nil
}
