package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
)

func (c *SessionHandler) SessionSendDataToSession(in *sessionpb.TSessionSendDataToSession) (*tproto.Bool, error) {
	var (
		data = in.GetData()
	)

	if data == nil {
		err := tproto.ErrInputRequestInvalid
		c.Logger.Errorf("session.sendDataToSession - error: %v", err)
		return nil, err
	}

	mainAuth, err := c.getOrFetchMainAuthWrapper(data.AuthId)
	if err != nil {
		c.Logger.Errorf("session.sendDataToSession - error: %v", err)
		return nil, err
	}

	_ = mainAuth.SessionDataArrived(
		c.ctx,
		data.AuthId,
		data.ServerId,
		data.ClientIp,
		data.SessionId,
		data.Payload)

	return tproto.BoolTrue, nil
}
