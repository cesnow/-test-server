package main

import (
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/test"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func main() {

	tproto.InitChecksum()

	var err error
	var c *websocket.Conn
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "localhost:7824", Path: "/"}
	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("WebSocket dial failed:", err)
	}
	test.MainClient = test.NewClient(c)
	defer test.MainClient.Close()

	done := make(chan bool)
	go test.Reader(done)

	test.SendConnCodecAndPing()

	<-test.WaitAuthReady
	_, _ = color.RGB(180, 159, 0).Print(
		"－－－－－－－－－－－－－－ Auth Ready －－－－－－－－－－－－－－\n")

	test.SendFakeUser(90001)

	go test.SendSchedule(30, 5000, test.SendPing)
	go test.SendSchedule(5, 5000, func() { test.SendTryNotifyToUser(90002, "test----A") })
	go test.SendSchedule(10, 5000, func() { test.SendTryNotifyToUser(90002, "test----B") })

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Interrupt signal received, closing connection...")
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
