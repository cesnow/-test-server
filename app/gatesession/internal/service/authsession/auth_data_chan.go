package authsession

import (
	"context"
)

type connDataCtx struct {
	ctx context.Context
	connData
}

type connData struct {
	isNew     bool
	authId    int64
	sessionId int64
	gatewayId string
}

type sessionDataCtx struct {
	ctx context.Context
	sessionData
}

type sessionData struct {
	authId    int64
	gatewayId string
	clientIp  string
	sessionId int64
	buf       []byte
}
type syncDataCtx struct {
	ctx context.Context
	syncData
}

type syncData struct {
	data *messageData
}
