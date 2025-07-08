package gnet

import (
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/codec"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/sse"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet/websocket"
)

type ConnectionType int

const (
	ConnectionTypeWebSocket ConnectionType = iota
	ConnectionTypeSSE
)

type connContext struct {
	connType  ConnectionType
	codec     codec.Codec
	authId    int64
	sessionId int64
	clientIp  string
	sseCodec  *sse.Codec
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
