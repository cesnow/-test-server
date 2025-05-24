package tproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var BoolTrue = &Bool{Value: true}
var BoolFalse = &Bool{Value: false}

func FromBool(b *Bool) bool {
	return b.Value
}

type TConstructor int32
type TObject proto.Message

func IsAnyOfType(anyMsg *anypb.Any, msg proto.Message) bool {
	return anyMsg.TypeUrl == "type.googleapis.com/"+string(msg.ProtoReflect().Descriptor().FullName())
}

func ConvertBytesToTMessage(msg []byte) (*TMessage, error) {
	var err error
	objLength := int32(binary.LittleEndian.Uint32(msg[12:]))
	obj := msg[16 : 16+objLength]

	classId := TConstructor(binary.LittleEndian.Uint32(obj[0:]))
	tObj := NewTObject(classId)
	if tObj == nil {
		return nil, errors.New(fmt.Sprintf("class id not found: %d", classId))
	}
	err = proto.Unmarshal(obj[4:], tObj)
	if err != nil {
		return nil, err
	}
	objPb, err := anypb.New(tObj)
	if err != nil {
		return nil, err
	}

	return &TMessage{
		MsgId:  int64(binary.LittleEndian.Uint64(msg[0:])),
		SeqNo:  int32(binary.LittleEndian.Uint32(msg[8:])),
		Bytes:  objLength,
		Object: objPb,
	}, err
}

func TryGetUnknownTObject(b []byte) (rList []*anypb.Any) {
	var (
		err      error
		messages []*TMessage
	)

	tMessage, err := ConvertBytesToTMessage(b)
	if err != nil {
		return
	}

	if ok := tMessage.Object.MessageIs(&TMsgContainer{}); ok {
		msgContainer := new(TMsgContainer)
		err = tMessage.Object.UnmarshalTo(msgContainer)
		if err != nil {
			return
		}
		messages = msgContainer.Messages
	} else {
		messages = append(messages, tMessage)
	}

	for _, m2 := range messages {
		rList = append(rList, m2.Object)
	}

	return
}

func EncodeTObject(obj TObject) []byte {
	buf := make([]byte, 4)
	checksum := GetTProtoChecksum(obj)
	binary.LittleEndian.PutUint32(buf[0:], uint32(checksum))
	payload, _ := proto.Marshal(obj)
	return append(buf, payload...)
}

func (tm *TMessage) Encode() []byte {
	actualTObj, _ := anypb.UnmarshalNew(tm.Object, proto.UnmarshalOptions{})
	checksum := GetTProtoChecksum(actualTObj)
	msgBytes, _ := proto.Marshal(actualTObj)

	buf := make([]byte, 20)
	binary.LittleEndian.PutUint64(buf[0:], uint64(tm.MsgId))
	binary.LittleEndian.PutUint32(buf[8:], uint32(tm.SeqNo))
	binary.LittleEndian.PutUint32(buf[12:], uint32(len(msgBytes)+4))
	binary.LittleEndian.PutUint32(buf[16:], uint32(checksum))

	return append(buf, msgBytes...)
}

func (mrd *MsgRawData) Encode() []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:], uint64(mrd.MsgId))
	binary.LittleEndian.PutUint32(buf[8:], uint32(mrd.SeqNo))
	binary.LittleEndian.PutUint32(buf[12:], uint32(mrd.Bytes))
	return append(buf, mrd.Body...)
}
