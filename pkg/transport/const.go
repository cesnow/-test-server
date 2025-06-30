package transport

import (
	"context"
	"time"
)

type TMsgRawData struct {
	MsgId int64 `json:"msgId" msgpack:"msgId"`
	//SeqNo    int32  `json:"seqNo" msgpack:"seqNo"`
	Event    string `json:"event" msgpack:"event"`
	ReqMsgId int64  `json:"reqMsgId,omitempty" msgpack:"reqMsgId,omitempty"`
	Body     []byte `json:"body" msgpack:"body"`
}

type TMsgRawDataContainer struct {
	Messages []*TMsgRawData `json:"messages"`
}

type TSendClientData struct {
	AuthId    int64
	SessionId int64
	GatewayId string
	Payload   TMsgRawData
}

type Pong struct {
	PingId int64 `msgpack:"pingId"`
}

type Metadata struct {
	Ctx          context.Context
	ServerId     string
	ClientAddr   string
	AuthId       int64
	SessionId    int64
	ReceivedTime time.Time
	UserId       int64
	ClientMsgId  int64
}
