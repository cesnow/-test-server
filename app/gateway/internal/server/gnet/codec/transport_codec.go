package codec

import (
	"encoding/binary"
	"errors"
)

const (
	MainFlag = 0x39F57B94
)

var (
	// ErrUnexpectedEOF occurs when no enough data to read by codec.
	ErrUnexpectedEOF = errors.New("there is no enough data")
)

type Reader interface {
	// Peek returns the next n bytes without advancing the reader. The bytes stop
	// being valid at the next read call. If Peek returns fewer than n bytes, it
	// also returns an error explaining why the read is short. The error is
	// ErrBufferFull if n is larger than b's buffer size.
	//
	// Note that the []byte buf returned by Peek() is not allowed to be passed to a new goroutine,
	// as this []byte will be reused within event-loop.
	// If you have to use buf in a new goroutine, then you need to make a copy of buf and pass this copy
	// to that new goroutine.
	Peek(n int) (buf []byte, err error)

	// Discard skips the next n bytes, returning the number of bytes discarded.
	//
	// If Discard skips fewer than n bytes, it also returns an error.
	// If 0 <= n <= b.Buffered(), Discard is guaranteed to succeed without
	// reading from the underlying io.Reader.
	Discard(n int) (discarded int, err error)
}

type Writer interface {
}

type Codec interface {
	Encode(conn Writer, msg interface{}) ([]byte, error)
	Decode(conn Reader) ([]byte, error)
}

func CreateCodec(conn Reader) (Codec, error) {
	var (
		firstInt uint32
	)

	firstBytes, err := conn.Peek(4)
	if err != nil {
		return nil, ErrUnexpectedEOF
	}

	firstInt = binary.BigEndian.Uint32(firstBytes)

	if firstInt == MainFlag {
		return NewMainCodec(), nil
	}

	return nil, errors.New("transportCodec header error")
}
