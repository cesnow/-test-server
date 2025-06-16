package gnet

import (
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/codec"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/websocket"
)

type connContext struct {
	codec     codec.Codec
	authId    int64
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
	return ctx.authId
}

func (ctx *connContext) putAuthId(authId int64) {
	ctx.authId = authId
}
