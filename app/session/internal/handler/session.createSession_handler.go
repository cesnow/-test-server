package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (c *SessionHandler) SessionCreateSession(in *sessionpb.TSessionCreateSession) (*tproto.Bool, error) {
	var (
		cli = in.GetClient()
	)

	if cli == nil {
		err := tproto.ErrInputRequestInvalid
		c.Logger.Errorf("session.createSession - error: %v", err)
		return nil, err
	}

	mainAuth, err := c.getOrFetchMainAuthWrapper(cli.AuthId)
	if err != nil {
		c.Logger.Errorf("session.createSession - error: %v", err)
		return nil, err
	}

	_ = mainAuth.SessionClientNew(c.ctx, cli.AuthId, cli.ServerId, cli.SessionId)

	return tproto.BoolTrue, nil
}
