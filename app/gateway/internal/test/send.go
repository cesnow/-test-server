package test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"google.golang.org/protobuf/proto"
	"kiyudesign.com/cesnow/light-server/pkg/atomic"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"time"
)

func getSeq() int32 {
	return seqNum.Add(1)
}

var msgIdSeq = atomic.NewAtomicInt64(0)

func nextMessageId2(req bool) int64 {
	unixNano := time.Now().UnixNano()
	ts := unixNano / 1e9
	ms := (unixNano % 1e9) / 1e6
	sid := msgIdSeq.Add(1) & 0x1ffff
	msgIdSeq.CompareAndSwap(0x1ffff, 0)
	last := 1
	if !req {
		last = 3
	}
	msgId := ts<<32 | int64(ms)<<21 | sid<<3 | int64(last)
	return msgId
}

var lastOutgoingMessageId = atomic.NewAtomicInt64(0)

func nextMessageId() int64 {
	const nano = 1000 * 1000 * 1000
	unixNano := time.Now().UnixNano()
	messageId := ((unixNano / nano) << 32) | ((unixNano % nano) & -4)
	if messageId <= lastOutgoingMessageId.Get() {
		messageId = lastOutgoingMessageId.Add(1)
	}
	for {
		if (messageId % 4) != 0 {
			messageId += 1
		} else {
			break
		}
	}
	lastOutgoingMessageId.Set(messageId)

	return messageId
}

func sendPackets(packets []byte, isAck bool) {
	if !isAck {
		writeTime = time.Now().UnixMilli()
	}
	writeCount += 1
	color.Green(fmt.Sprintf("█ SendPacket [%s]",
		hex.EncodeToString(packets)))
	MainClient.Send(packets)
}

func makeMsgData(message tproto.TObject) []byte {

	checksum := tproto.GetTProtoChecksum(message)
	messageBytes, _ := proto.Marshal(message)
	checksumBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(checksumBytes[0:], uint32(checksum)) // checksum
	payloadBytes := append(checksumBytes, messageBytes...)

	buf := make([]byte, 16)
	nextMsgId := nextMessageId()
	binary.LittleEndian.PutUint64(buf[0:], uint64(nextMsgId))          // msg_id
	binary.LittleEndian.PutUint32(buf[8:], uint32(getSeq()))           // seqNo
	binary.LittleEndian.PutUint32(buf[12:], uint32(len(payloadBytes))) // payload_length

	buf = append(buf, payloadBytes...)

	color.Yellow(fmt.Sprintf(
		"－－－－－－－－－－－－－－－ Make Send Msg (%s, %d) －－－－－－－－－－－－－－ ",
		message.ProtoReflect().Descriptor().FullName(), nextMsgId))
	printFmt := "%-18s %-10s %-16s %-10s %s"
	fmt.Println(fmt.Sprintf(printFmt, "MSG_ID", "SEQ_NO", "PAYLOAD_SIZE", "CLASS_ID", "PAYLOAD"))
	fmt.Println(fmt.Sprintf(printFmt,
		hex.EncodeToString(buf[0:8]),
		hex.EncodeToString(buf[8:12]),
		hex.EncodeToString(buf[12:16]),
		hex.EncodeToString(buf[16:20]),
		hex.EncodeToString(buf[24:]),
	))

	return buf
}

func makePacketData(msgPayload []byte) []byte {
	dataBuf := make([]byte, 8+8)
	binary.LittleEndian.PutUint64(dataBuf[0:], uint64(mainAuthId))
	binary.LittleEndian.PutUint64(dataBuf[8:], uint64(mainSessionId))
	dataBuf = append(dataBuf, msgPayload...)

	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(len(dataBuf)))

	return append(sizeBuf, dataBuf...)
}
