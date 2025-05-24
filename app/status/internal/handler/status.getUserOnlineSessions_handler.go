package handler

import (
	"github.com/zeromicro/go-zero/core/jsonx"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
)

func (c *StatusCore) StatusGetUserOnlineSessions(in *statuspb.TStatusGetUserOnlineSessions) (*statuspb.UserSessionEntryList, error) {

	rMapCmd := c.svcCtx.KV.B().Hgetall().Key(getUserKey(in.GetUserId())).Build()
	rMap, err := c.svcCtx.KV.Do(c.ctx, rMapCmd).AsStrMap()

	if err != nil {
		c.Logger.Errorf("status.getUserOnlineSessions(%s) error(%v)", in, err)
		return nil, err
	}

	var (
		rValues = &statuspb.UserSessionEntryList{
			UserId:       in.UserId,
			UserSessions: make([]*statuspb.SessionEntry, 0, len(rMap)),
		}
	)

	for _, v := range rMap {
		sess := new(statuspb.SessionEntry)
		if err2 := jsonx.UnmarshalFromString(v, sess); err2 == nil {
			rValues.UserSessions = append(rValues.UserSessions, sess)
		}
	}

	return rValues, nil
}
