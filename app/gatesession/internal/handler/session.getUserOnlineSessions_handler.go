package handler

import (
	"context"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/service/authsession"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
)

type SessionGetUserOnlineSessionsRequest struct {
	UserId int64 `json:"userId"`
}

type SessionGetUserOnlineSessionsResponse struct {
	UserId            int64                    `json:"userId"`
	UserSessions      []*statuspb.SessionEntry `json:"userSessions"`
	ServerIdKeyIdList map[string][]int64       `json:"serverIdKeyIdList"`
}

func (h *Handler) SessionGetUserOnlineSessions(ctx context.Context, in *SessionGetUserOnlineSessionsRequest) (*SessionGetUserOnlineSessionsResponse, error) {

	rMap, err := h.svcCtx.KV.HgetallCtx(ctx, authsession.GetOnlineUserKey(in.UserId))
	if err != nil {
		logx.Errorf("session.getUserOnlineSessions(%d) error(%v)", in.UserId, err)
		return nil, err
	}

	var (
		rValues = &SessionGetUserOnlineSessionsResponse{
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

	var serverIdKeyIdList = make(map[string][]int64)

	logx.Debugf("statusList - #%v", rValues.UserSessions)

	for _, sess := range rValues.UserSessions {
		if keyIdList, ok := serverIdKeyIdList[sess.Gateway]; ok {
			keyIdList = append(keyIdList, sess.AuthId)
			serverIdKeyIdList[sess.Gateway] = keyIdList
		} else {
			serverIdKeyIdList[sess.Gateway] = []int64{sess.AuthId}
		}
	}

	rValues.ServerIdKeyIdList = serverIdKeyIdList

	return rValues, nil

}
