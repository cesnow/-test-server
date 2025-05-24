package test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"math/rand"
	"time"
)

func makePingPackets() []byte {
	msgData := makeMsgData(&tproto.Ping{PingId: rand.Int63()})
	msgPackets := makePacketData(msgData)
	return msgPackets
}

func makeHelpGetCountriesListPackets() []byte {
	msgData := makeMsgData(&configurationpb.THelpGetCountriesList{
		LangCode: "zh-TW",
		Hash:     0,
	})
	msgPackets := makePacketData(msgData)
	return msgPackets
}

func SendPing() {
	data := makePingPackets()
	sendPackets(data, false)
}

func SendFakeUser(userId int64) {
	msgData := makeMsgData(&sessionpb.TSession_TTTTT_TestUser{
		UserId: userId,
		Status: 1,
	})
	msgPackets := makePacketData(msgData)
	sendPackets(msgPackets, false)
}

func SendTryNotifyToUser(userId int64, msg string) {
	msgData := makeMsgData(&configurationpb.THelpTryNotifyUser{
		UserId:  userId,
		Message: msg,
	})
	msgPackets := makePacketData(msgData)
	sendPackets(msgPackets, false)
}

func SendHelpGetCountriesList() {
	data := makeHelpGetCountriesListPackets()
	sendPackets(data, false)
}

func SendSchedule(t time.Duration, limit int, f func()) {
	count := 0
	for {
		time.Sleep(time.Second * t)
		f()
		count += 1
		if count >= limit {
			break
		}
	}
}

func SendConnCodecAndPing() {
	buf := connCodecAndPing()
	sendPackets(buf, false)
}

func SendMsgAck(ackMsgIds []int64) {
	color.Cyan("█ SendAckIds: %v\n", ackMsgIds)
	ackData := makeMsgData(&tproto.TMsgAck{MsgIds: ackMsgIds})
	msgAckPackets := makePacketData(ackData)
	sendPackets(msgAckPackets, true)
}

func connCodecAndPing() []byte {

	pingPackets := makePingPackets()

	headerBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(headerBuf, uint32(0x39F57B94))

	color.Magenta("－－－－－－－－－－－－－－－－－－－－－－ FIRST CODEC －－－－－－－－－－－－－－－－－－－－－ ")
	fmt.Println("PEEK | SIZE | AUTH_ID | SESSION_ID | MSG_PAYLOAD")
	fmt.Println(fmt.Sprintf("%s | %s | %s | %s | %s ",
		hex.EncodeToString(headerBuf),
		hex.EncodeToString(pingPackets[0:4]),
		hex.EncodeToString(pingPackets[4:12]),
		hex.EncodeToString(pingPackets[12:20]),
		hex.EncodeToString(pingPackets[20:])))

	buf := append(headerBuf, pingPackets...)
	return buf
}
