package tproto

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"hash/crc32"
	"math"
)

var tProtoCrc map[protoreflect.FullName]TConstructor
var CrcTProto map[TConstructor]protoreflect.Message

func InitChecksum() {
	tProtoCrc = make(map[protoreflect.FullName]TConstructor)
	CrcTProto = make(map[TConstructor]protoreflect.Message)

	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		name := mt.Descriptor().FullName()
		checksum := crc32.ChecksumIEEE([]byte(name))
		var crc32Int TConstructor

		if checksum > math.MaxInt32 {
			crc32Int = TConstructor(checksum - math.MaxUint32 - 1)
		} else {
			crc32Int = TConstructor(checksum)
		}
		CrcTProto[crc32Int] = mt.New()
		tProtoCrc[name] = crc32Int
		logx.Debugf(fmt.Sprintf("crc[%d] is mapping to name [%s]", crc32Int, mt.Descriptor().FullName()))
		return true
	})
	logx.Infof("[tProtoInit] Loaded types: %d", len(tProtoCrc))
}

func GetTProtoChecksum(obj TObject) TConstructor {
	if checksum, find := tProtoCrc[obj.ProtoReflect().Descriptor().FullName()]; find {
		return checksum
	}
	return -1
}

func NewTObject(checksum TConstructor) TObject {
	if message, find := CrcTProto[checksum]; find {
		return message.Interface()
	}
	return nil
}
