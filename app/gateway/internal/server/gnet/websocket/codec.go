package websocket

import (
	"bytes"
	"errors"
	"github.com/panjf2000/gnet/v2"
	"io"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type Codec struct {
	upgraded bool
	Buf      bytes.Buffer
	wsMsgBuf wsMessageBuf
	Conn     Connection
}

type wsMessageBuf struct {
	curHeader *ws.Header
	cachedBuf bytes.Buffer
}

type readWrite struct {
	io.Reader
	io.Writer
}

func (w *Codec) Upgrade(c gnet.Conn) (ok bool, action gnet.Action) {
	if w.upgraded {
		ok = true
		return
	}
	buf := &w.Buf
	tmpReader := bytes.NewReader(buf.Bytes())
	oldLen := tmpReader.Len()

	upgrader := ws.Upgrader{
		Protocol: func(i []byte) bool {
			return bytes.Equal(i, []byte("binary"))
		},
		//ProtocolCustom: func(i []byte) (string, bool) {
		//	return "", true
		//},
		Extension: func(option httphead.Option) bool {
			return true
		},
		ExtensionCustom: func(i []byte, options []httphead.Option) ([]httphead.Option, bool) {
			return nil, true
		},
		//Negotiate: func(option httphead.Option) (httphead.Option, error) {
		//	return
		//},
		OnRequest: func(uri []byte) error {
			return nil
		},
		OnHost: func(host []byte) error {
			return nil
		},
		OnHeader: func(key, value []byte) error {
			return nil
		},
		//OnBeforeUpgrade: func() (header ws.HandshakeHeader, err error) {
		//	//
		//},
	}

	hs, err := upgrader.Upgrade(readWrite{tmpReader, c})
	skipN := oldLen - tmpReader.Len()
	if err != nil {
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) { //数据不完整，不跳过 buf 中的 skipN 字节（此时 buf 中存放的仅是部分 "handshake data" bytes），下次再尝试读取
			return
		}
		buf.Next(skipN)
		logx.Errorf("conn[%v] req[%s][err=%v]", c.RemoteAddr().String(), buf.String(), err.Error())
		action = gnet.Close
		return
	}
	buf.Next(skipN)
	logx.Debugf("conn[%v] upgrade websocket! Handshake: %v", c.RemoteAddr().String(), hs)

	ok = true
	w.upgraded = true
	return
}

func (w *Codec) ReadBufferBytes(c gnet.Conn) gnet.Action {
	size := c.InboundBuffered()
	buf := make([]byte, size)
	read, err := c.Read(buf)
	if err != nil {
		logx.Errorf("read err! %v", err)
		return gnet.Close
	}
	if read < size {
		logx.Errorf("read bytes len err! size: %d read: %d", size, read)
		return gnet.Close
	}
	w.Buf.Write(buf)
	return gnet.None
}

func (w *Codec) Decode(c gnet.Conn) (outs []wsutil.Message, err error) {
	logx.Debug("do Decode")
	messages, err := w.readWsMessages()
	if err != nil {
		logx.Errorf("Error reading message! %v", err)
		return nil, err
	}
	if messages == nil || len(messages) <= 0 { //没有读到完整数据 不处理
		return
	}
	for _, message := range messages {
		if message.OpCode.IsControl() {
			err = wsutil.HandleClientControlMessage(c, message)
			if err != nil {
				return
			}
			continue
		}
		if message.OpCode == ws.OpText || message.OpCode == ws.OpBinary {
			outs = append(outs, message)
		}
	}
	return
}

func (w *Codec) readWsMessages() (messages []wsutil.Message, err error) {
	msgBuf := &w.wsMsgBuf
	in := &w.Buf
	for {
		// read header from `in` and write header bytes to msgBuf.cachedBuf.
		if msgBuf.curHeader == nil {
			if in.Len() < ws.MinHeaderSize {
				return
			}
			var head ws.Header
			if in.Len() >= ws.MaxHeaderSize {
				head, err = ws.ReadHeader(in)
				if err != nil {
					return messages, err
				}
			} else {
				tmpReader := bytes.NewReader(in.Bytes())
				oldLen := tmpReader.Len()
				head, err = ws.ReadHeader(tmpReader)
				skipN := oldLen - tmpReader.Len()
				if err != nil {
					if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) { //数据不完整
						return messages, nil
					}
					in.Next(skipN)
					return nil, err
				}
				in.Next(skipN)
			}

			msgBuf.curHeader = &head
			err = ws.WriteHeader(&msgBuf.cachedBuf, head)
			if err != nil {
				return nil, err
			}
		}
		dataLen := (int)(msgBuf.curHeader.Length)

		if dataLen > 0 {
			if in.Len() < dataLen {
				logx.Debug(in.Len(), dataLen)
				logx.Infof("incomplete data")
				return
			}

			_, err = io.CopyN(&msgBuf.cachedBuf, in, int64(dataLen))
			if err != nil {
				return
			}
		}
		if msgBuf.curHeader.Fin {
			messages, err = wsutil.ReadClientMessage(&msgBuf.cachedBuf, messages)
			if err != nil {
				return nil, err
			}
			msgBuf.cachedBuf.Reset()
		} else {
			logx.Infof("The data is split into multiple frames")
		}
		msgBuf.curHeader = nil
	}
}
