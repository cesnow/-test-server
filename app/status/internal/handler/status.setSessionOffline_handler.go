package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
	"strconv"
)

func (c *StatusCore) StatusSetSessionOffline(in *statuspb.TStatusSetSessionOffline) (*tproto.Bool, error) {

	delBuild := c.svcCtx.Service.KV.B().Hdel().Key(getUserKey(in.GetUserId())).Field(strconv.FormatInt(in.GetAuthId(), 10)).Build()
	err := c.svcCtx.Service.KV.Do(c.ctx, delBuild).Error()

	if err != nil {
		c.Logger.Errorf("status.setSessionOffline(%s) error(%v)", in, err)
		return nil, err
	}

	return tproto.BoolTrue, nil
}
