package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"math/rand"
)

func (c *SessionHandler) SessionQueryAuthId(in *sessionpb.TSessionQueryAuthId) (*tproto.AuthIdInfo, error) {

	// call AuthSessionClient to get new one
	// mock first

	newAuth := &tproto.AuthIdInfo{
		AuthId:    rand.Int63(),
		SessionId: rand.Int63(),
		AuthKey:   []byte{},
	}

	return newAuth, nil
}
