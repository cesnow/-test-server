package gnet

import (
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/codec"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/websocket"
)

type AuthIdInfo struct {
	AuthId    int64
	SessionId int64
}

type connContext struct {
	codec     codec.Codec
	authInfo  *AuthIdInfo
	sessionId int64
	clientIp  string
	wsCodec   *websocket.Codec
	logx.Logger
	closeDate int64
}

func newConnContext() *connContext {
	return &connContext{
		codec:    codec.NewMainCodec(),
		clientIp: "",
	}
}

func (ctx *connContext) setClientIp(ip string) {
	ctx.clientIp = ip
}

func (ctx *connContext) AuthId() int64 {
	if ctx.authInfo == nil {
		return -1
	}
	return ctx.authInfo.AuthId
}

func (ctx *connContext) putAuthIdInfo(authIdInfo *AuthIdInfo) {
	ctx.authInfo = authIdInfo
}

func (ctx *connContext) getAuthIdInfo() *AuthIdInfo {
	return ctx.authInfo
}
