package transport

type TMsgRawData struct {
	MsgId int64  `json:"msgId" msgpack:"msgId"`
	SeqNo int32  `json:"seqNo" msgpack:"seqNo"`
	Event string `json:"event" msgpack:"event"`
	Body  []byte `json:"body" msgpack:"body"`
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
