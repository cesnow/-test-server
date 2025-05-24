package main

import (
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

	go test.SendSchedule(10, 1000, test.SendPing)
	go test.SendSchedule(15, 10, test.SendHelpGetCountriesList)

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
