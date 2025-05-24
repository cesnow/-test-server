package main

import (
	"encoding/hex"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type Item struct {
	MsgId int64 `msgpack:"msgId"`
}

func main() {

	b, err := msgpack.Marshal(&Item{MsgId: 15}) //9223372036854775807
	if err != nil {
		panic(err)
	}

	logx.Infof("%s", hex.EncodeToString(b))

	var item Item
	err = msgpack.Unmarshal(b, &item)
	if err != nil {
		panic(err)
	}
	fmt.Println(item.MsgId)
	// 81
	//  a5   6d 73 67 49 64
	// d3    000000000000000f

	// 81a56d73674964d700000000000000000f

	// 81a56d73674964 d3000000000000000f
	// 81a56d73674964 cf000000000000000f

	// 81a56d73674964 d4000f
}
