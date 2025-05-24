package gnet

import (
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/gnet/codec"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/gnet/websocket"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
)

type connContext struct {
	codec     codec.Codec
	authInfo  *tproto.AuthIdInfo
	sessionId int64
	clientIp  string
	wsCodec   *websocket.Codec
	logx.Logger
	closeDate int64
}

func newConnContext() *connContext {
	return &connContext{
		codec:    nil,
		clientIp: "",
	}
}

func (ctx *connContext) setClientIp(ip string) {
	ctx.clientIp = ip
}

func (ctx *connContext) AuthId() int64 {
	return ctx.authInfo.AuthId
}

func (ctx *connContext) putAuthIdInfo(authIdInfo *tproto.AuthIdInfo) {
	ctx.authInfo = authIdInfo
}

func (ctx *connContext) getAuthIdInfo() *tproto.AuthIdInfo {
	return ctx.authInfo
}
