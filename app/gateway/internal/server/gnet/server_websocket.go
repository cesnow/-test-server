package gnet

import (
	"encoding/binary"
	"errors"
	"github.com/panjf2000/gnet/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server/gnet/codec"
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

		if ctx.codec == nil {
			ctx.codec, err = codec.CreateCodec(&ctx.wsCodec.Conn)
			if err != nil {
				if errors.Is(err, codec.ErrUnexpectedEOF) {
					return gnet.None
				}
				logx.Errorf("conn(%d-%s) create codec error: %v", c.Fd(), c.RemoteAddr().String(), err)
				return gnet.Close
			}
			_, _ = ctx.wsCodec.Conn.Discard(4)
		}

		frame, err := ctx.codec.Decode(&ws.Conn)
		if err != nil {
			logx.Errorf("conn(%d-%s) frame is error: %v", c.Fd(), c.RemoteAddr().String(), err)
			action = gnet.Close
			return
		} else if frame == nil {
			logx.Debugf("conn(%d-%s) frame is nil", c.Fd(), c.RemoteAddr().String())
			return
		}

		action = s.onReceiveRawMessage(ctx, c, int64(binary.LittleEndian.Uint64(frame)), frame)
		if action == gnet.Close {
			return
		}

		_, _ = ws.Conn.InboundBuffer.Write(ws.Conn.Buffer)
		ws.Conn.Buffer = ws.Conn.Buffer[:0]

	}
	return gnet.None
}
