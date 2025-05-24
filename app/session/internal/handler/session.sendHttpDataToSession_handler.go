package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"time"
)

func (c *SessionHandler) SessionSendHttpDataToSession(in *sessionpb.TSessionSendHttpDataToSession) (*sessionpb.HttpSessionData, error) {
	var (
		data = in.GetClient()
	)

	if data == nil {
		err := tproto.ErrInputRequestInvalid
		c.Logger.Errorf("session.sendHttpDataToSession - error: %v", err)
		return nil, err
	}

	mainAuth, err := c.getOrFetchMainAuthWrapper(data.AuthId)
	if err != nil {
		c.Logger.Errorf("session.sendHttpDataToSession - error: %v", err)
		return nil, err
	}

	chData := make(chan interface{})
	_ = mainAuth.SessionHttpDataArrived(
		c.ctx,
		data.AuthId,
		data.ServerId,
		data.ClientIp,
		data.SessionId,
		data.Payload,
		chData)

	timer := time.NewTimer(time.Second * 7)
	select {
	case cData := <-chData:
		return &sessionpb.HttpSessionData{
			Payload: cData.([]byte),
		}, nil
	case <-timer.C:
		c.Logger.Errorf("chData timeout...")
	}

	return &sessionpb.HttpSessionData{
		Payload: []byte{},
	}, nil
}
