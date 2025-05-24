package gnet

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Server) onWebsocketData(ctx *connContext, c gnet.Conn) (action gnet.Action) {
	ws := ctx.wsCodec
	if ws.ReadBufferBytes(c) == gnet.Close {
		return gnet.Close
	}
	ok, action := ws.Upgrade(c)
	if !ok {
		return
	}

	if ws.Buf.Len() <= 0 {
		return gnet.None
	}
	messages, err := ws.Decode(c)
	if err != nil {
		return gnet.Close
	}
	if messages == nil {
		return
	}
	for _, message := range messages {
		ws.Conn.Buffer = message.Payload

		frame, err := ctx.codec.Decode(&ws.Conn)
		if err != nil {
			logx.Errorf("conn(%d-%s) frame is error: %v", c.Fd(), c.RemoteAddr().String(), err)
			action = gnet.Close
			return
		} else if frame == nil {
			logx.Debugf("conn(%d-%s) frame is nil", c.Fd(), c.RemoteAddr().String())
			return
		}

		action = s.onReceiveRawMessage(ctx, c, frame)
		if action == gnet.Close {
			return
		}

		_, _ = ws.Conn.InboundBuffer.Write(ws.Conn.Buffer)
		ws.Conn.Buffer = ws.Conn.Buffer[:0]

	}
	return gnet.None
}
