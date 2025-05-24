package codec

import (
	"encoding/binary"
	"fmt"
)

type MainCodec struct{}

func NewMainCodec() *MainCodec {
	return new(MainCodec)
}

func (c *MainCodec) Decode(conn Reader) ([]byte, error) {

	var (
		size int
		buf  []byte
		n    int
		in   innerBuffer
		err  error
	)

	in, _ = conn.Peek(-1)

	if buf, err = in.readN(4); err != nil {
		return nil, ErrUnexpectedEOF
	}
	size += 4

	n = int(binary.LittleEndian.Uint32(buf))

	if n < 24 {
		err = fmt.Errorf("invalid len: %d", size)
		return nil, err
	}

	if buf, err = in.readN(n); err != nil {
		return nil, ErrUnexpectedEOF
	}
	size += n
	_, _ = conn.Discard(size)

	return buf, nil
	// allSize[4] ( authId[8], sessionId[8], msgId[8], seqNo[4], Bytes[4], Payload[Class[4], Data...] )
}

func (c *MainCodec) Encode(conn Writer, msg any) ([]byte, error) {

	payload := msg.([]byte)

	// allSize[4] ( authId[8], sessionId[8], msgId[8], seqNo[4], Bytes[4], Payload[Class[4], Data...] )

	b := payload

	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(len(b)))

	b = append(sizeBytes, b...)

	return b, nil
}
