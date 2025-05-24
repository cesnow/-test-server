package handler

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
	"strconv"

	"github.com/zeromicro/go-zero/core/jsonx"
)

func (c *StatusCore) StatusSetSessionOnline(in *statuspb.TStatusSetSessionOnline) (*tproto.Bool, error) {
	var (
		userK = getUserKey(in.GetUserId())
		sess  = in.GetSession()
	)

	sessData, _ := jsonx.Marshal(sess)
	setBuild := c.svcCtx.Service.KV.B().Hset().Key(userK).FieldValue().FieldValue(
		strconv.FormatInt(sess.GetAuthId(), 10), string(sessData),
	).Build()
	err := c.svcCtx.Service.KV.Do(c.ctx, setBuild).Error()
	if err != nil {
		c.Logger.Errorf("status.setSessionOnline(%s) error(%v)", in, err)
		return nil, err
	}

	expireBuild := c.svcCtx.Service.KV.B().Expire().Key(userK).Seconds(int64(c.svcCtx.Config.StatusExpire)).Build()
	err = c.svcCtx.Service.KV.Do(c.ctx, expireBuild).Error()
	if err != nil {
		c.Logger.Errorf("status.setSessionOnline(%s) error(%v)", in, err)
		return nil, err
	}

	return tproto.BoolTrue, nil
}
