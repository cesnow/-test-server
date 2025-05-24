package codec

type MainCodec struct{}

func NewMainCodec() *MainCodec {
	return new(MainCodec)
}

func (c *MainCodec) Decode(conn Reader) ([]byte, error) {

	var (
		//	size int
		//	buf  []byte
		//	n    int
		in innerBuffer
		//	err  error
	)

	in, _ = conn.Peek(-1)

	_, _ = conn.Discard(len(in))

	return in, nil

}

func (c *MainCodec) Encode(conn Writer, msg any) ([]byte, error) {

	payload := msg.([]byte)

	return payload, nil
}
