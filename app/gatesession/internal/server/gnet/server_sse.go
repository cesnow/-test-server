package gnet

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

func (s *Server) onSSEData(ctx *connContext, c gnet.Conn) (action gnet.Action) {
	// without received
	buffer, err := c.Next(-1)

	logx.Infof("SSE received data: %s", string(buffer))
	logx.Infof("SSE err: %v", err)

	if err != nil {
		return gnet.Close
	}

	// s.onReceiveSSEMessage(ctx, c, data)
	//if s.isSSERequest(c) {
	ctx.connType = ConnectionTypeSSE

	headers := "HTTP/1.1 200 OK\r\n"
	headers += "Content-Type: text/event-stream\r\n"
	headers += "Cache-Control: no-cache\r\n"
	headers += "Connection: keep-alive\r\n"
	headers += "Access-Control-Allow-Origin: *\r\n"
	headers += "\r\n"

	out := []byte(headers)
	//
	initialEvent, _ := ctx.sseCodec.Encode("connected", "Connection established")
	out = append(out, initialEvent...)
	//}
	_, _ = c.Write(out)

	return gnet.None
}

func (s *Server) sendSSEHeartbeat() {
	s.Iterate(func(c gnet.Conn) {
		ctx, _ := c.Context().(*connContext)
		if ctx != nil && ctx.connType == ConnectionTypeSSE {
			heartbeat, _ := ctx.sseCodec.Encode("heartbeat", time.Now().Unix())
			_, _ = c.Write(heartbeat)
		}
	})
}

//func (s *Server) isSSERequest(c gnet.Conn) bool {
//	buffer, err := c.Next(-1)
//	logx.Infof("%+v", string(buffer))
//	logx.Infof("%+v", err)
//	if err != nil {
//		return false
//	}
//	req := string(buffer)
//
//	logx.Infof("%+v", req)
//
//	if strings.HasPrefix(req, "GET /sse") {
//		headers := "HTTP/1.1 200 OK\r\n" +
//			"Content-Type: text/event-stream\r\n" +
//			"Cache-Control: no-cache\r\n" +
//			"Connection: keep-alive\r\n\r\n"
//		_, _ = c.Write([]byte(headers))
//		return true
//	}
//
//	return false
//}
