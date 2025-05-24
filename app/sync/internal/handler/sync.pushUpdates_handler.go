package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
)

func (c *SyncHandler) SyncPushUpdates(in *syncpb.TSyncPushUpdates) (*tproto.Void, error) {
	var (
		userId  = in.GetUserId()
		updates = in.GetUpdates()
	)

	c.pushUpdatesToSession(userId, nil, nil, nil, updates)

	return &tproto.Void{}, nil
}
