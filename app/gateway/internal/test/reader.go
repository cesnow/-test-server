package test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"log"
	"time"
)

func Reader(done chan bool) {
	// wait for reply
	defer close(done)
	for {
		typo, message, err := MainClient.Reader()
		if err != nil {
			log.Println("Failed to read WebSocket message:", err)
			return
		}
		receivedCount += 1
		decodeMsg(message, typo, len(message))
	}
}

func decodeMsg(msg []byte, typo, n int) {

	diffTime := int64(0.0)
	diffMsg := "Active"
	if writeTime > 0 {
		diffTime = time.Now().UnixMilli() - writeTime
		writeTime = 0.0
		diffMsg = fmt.Sprintf("%d ms", diffTime)
	}

	color.Blue(fmt.Sprintf(
		"－－－－－－－－－－－－－－－－－－－－－－ Received Msg －－－－－－－－－－－－－－－－－－－－－ %s",
		diffMsg,
	))
	fmt.Printf("size: %d [length data: %s] \n", n, hex.EncodeToString(msg[:4]))

	size := int(binary.LittleEndian.Uint32(msg[0:]))

	if size > len(msg)-4 {
		color.Red("Size failed")
		return
	}

	payload := msg[4 : size+4]
	fmt.Println(fmt.Sprintf("size: %d, payload: %s", size, hex.EncodeToString(payload)))

	//authId = int64(binary.LittleEndian.Uint64(payload[0:]))
	sessionId := int64(binary.LittleEndian.Uint64(payload[8:]))
	tMessage, err := tproto.ConvertBytesToTMessage(payload[16:])
	if err != nil {
		color.Red(fmt.Sprintf("convert to TMessage failed, %s", err.Error()))
		return
	}

	fmt.Println(fmt.Sprintf(
		"AuthId: %d, SessionId: %d, SeqNumber: %d, MsgId: %d",
		mainAuthId, sessionId, tMessage.SeqNo, tMessage.MsgId))

	if tMessage.MsgId <= lastReadMessageId {
		return
	}
	lastReadMessageId = tMessage.MsgId

	tObj, err := anypb.UnmarshalNew(tMessage.Object, proto.UnmarshalOptions{})
	if err != nil {
		return
	}
	ackMsgIds := parseObject(tMessage.MsgId, sessionId, tObj)

	if len(ackMsgIds) != 0 {
		SendMsgAck(ackMsgIds)
	}

	if n > size+4 {
		time.Sleep(time.Millisecond * 100)
		color.Red("ContinueDecodeMsg")
		unProcessPayload := msg[size+4:]
		decodeMsg(unProcessPayload, typo, n-(size+4))
	}

	printLine()

	return
}

func parseObject(msgId, sessionId int64, msg proto.Message) []int64 {
	var ackIds []int64
	var subAckIds []int64
	switch request := msg.(type) {
	case *anypb.Any:
		unmarshalObj, err := request.UnmarshalNew()
		if err != nil {
			return ackIds
		}
		subAckIds = parseObject(msgId, sessionId, unmarshalObj)
	case *tproto.AuthIdInfo:
		mainAuthId = request.AuthId
		mainSessionId = request.SessionId
		color.White("█ AuthIdInfo -> AuthId: %d, SessionId: %d", request.AuthId, request.SessionId)
		WaitAuthReady <- true
	case *tproto.NewSessionCreated:
		color.White("█ NewSessionCreated -> MsgId: %d, Data: %+v", msgId, request.String())
		ackIds = append(ackIds, msgId)
	case *tproto.Pong:
		color.White("█ Pong -> pingId: %d", request.GetPingId())
	case *tproto.RpcResult:
		fmt.Printf("%+v\n", request.Result)
		color.White("█ RpcResult -> %v", request.Result.String())
		subAckIds = parseObject(msgId, sessionId, request.Result)
		ackIds = append(ackIds, msgId)
	case *tproto.RpcError:
		color.White("█ RpcError -> (%d) %s", request.GetErrorCode(), request.GetErrorMessage())
	case *tproto.TMsgRawDataContainer:
		for _, msg2 := range request.Messages {
			classId := tproto.TConstructor(binary.LittleEndian.Uint32(msg2.Body[0:]))
			tObj := tproto.NewTObject(classId)
			if tObj == nil {
				_, _ = color.RGB(233, 190, 231).Printf("CONVERT_TOBJECT_ERR_CLASS_ID: %d \n", classId)
				continue
			}
			parseObject(msg2.MsgId, sessionId, tObj)
			subAckIds = append(subAckIds, msg2.MsgId)
		}
	case *tproto.TMsgContainer:
		for _, msg2 := range request.Messages {
			// fmt.Println(fmt.Sprintf("ReceivedContainerMsg: Id: %d, Debug: %+v, Obj: %s", msgId, msg2.String(), msg2.Object.String()))
			parseObject(msg2.MsgId, sessionId, msg2.Object)
			subAckIds = append(subAckIds, msg2.MsgId)
		}
	case *tproto.TGzipPacked:
		color.White("█ GzipPacked -> %s", request.String())
		parseObject(msgId, sessionId, request.Obj)
	default:
		_, _ = color.RGB(190, 190, 190).Printf(
			"█ [TOBJECT_NOT_HANDLED] MsgId: %d, <%s> {{{ %v }}} \n",
			msgId, msg.ProtoReflect().Descriptor().FullName(), msg)
		ackIds = append(ackIds, msgId)
	}

	return append(ackIds, subAckIds...)
}

func printLine() {
	color.Cyan("－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－－")
}
