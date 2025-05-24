package gnet

import (
	"encoding/binary"
	"kiyudesign.com/cesnow/light-server/pkg/atomic"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"time"
)

var msgIdSeq = atomic.NewAtomicInt64(0)

func nextMessageId() int64 {
	unixNano := time.Now().UnixNano()
	ts := unixNano / 1e9
	ms := (unixNano % 1e9) / 1e6
	sid := msgIdSeq.Add(1) & 0x1ffff
	msgIdSeq.CompareAndSwap(0x1ffff, 0)
	last := 1
	msgId := ts<<32 | int64(ms)<<21 | sid<<3 | int64(last)
	return msgId
}

func serializeToBuffer(authId, sessionId int64, message *tproto.TMessage) []byte {
	return serializeUserToBuffer(authId, sessionId, message.Encode())
}

func serializeUserToBuffer(authId, sessionId int64, payload []byte) []byte {
	userBytes := make([]byte, 16)
	binary.LittleEndian.PutUint64(userBytes[0:], uint64(authId))
	binary.LittleEndian.PutUint64(userBytes[8:], uint64(sessionId))
	return append(userBytes, payload...)
}
