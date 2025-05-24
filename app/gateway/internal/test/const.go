package test

import (
	"github.com/gorilla/websocket"
	"kiyudesign.com/cesnow/light-server/pkg/atomic"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
	done chan struct{}
}

var (
	MainClient        *Client
	mainAuthId        int64              = 0
	mainSessionId     int64              = 0
	seqNum            atomic.AtomicInt32 = atomic.NewAtomicInt32(0)
	writeTime         int64              = 0
	writeCount        int64              = 0
	receivedCount     int64              = 0
	WaitAuthReady                        = make(chan bool)
	lastReadMessageId int64              = 0
)

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		send: make(chan []byte, 512),
		done: make(chan struct{}),
	}
	go c.writerPump()
	return c
}

func (c *Client) Reader() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *Client) writerPump() {
	defer c.conn.Close()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				return
			}
		case <-c.done:
			_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
	}
}

func (c *Client) Close() {
	close(c.done)
	close(c.send)
}

func SetSessionId(id int64) {
	mainSessionId = id
}
