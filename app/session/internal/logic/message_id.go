package logic

import (
	"kiyudesign.com/cesnow/light-server/pkg/atomic"
	"time"
)

var msgIdSeq = atomic.NewAtomicInt64(0)

func nextMessageId() int64 {
	unixNano := time.Now().UnixNano()
	ts := unixNano / 1e9
	ms := (unixNano % 1e9) / 1e6
	sid := msgIdSeq.Add(1) & 0x1ffff
	msgIdSeq.CompareAndSwap(0x1ffff, 0)
	msgId := ts<<32 | int64(ms)<<21 | sid<<3 | int64(1)
	return msgId
}
