package remote

import (
	"bytes"
	"log"
)

func encodeWrite(fd int64, b []byte) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.EncodeCall(SYS_WRITE, fd, b); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeWrite(m *Message, fd *int64, b *[]byte) error {
	return m.Decode(fd, b)
}

func (rh *RemoteHost) write(fd int64, b []byte) (n int, err error) {

	// Encode
	call, err := encodeWrite(fd, b)
	if err != nil {
		return 0, err
	}

	// Send
	m, err := rh.sendCall2(call)
	if err != nil {
		return 0, err
	}

	// Handle reply
	if err := m.Decode(&n); err != nil {
		return 0, err
	}

	return n, nil
}

func (lh *LocalHost) write(m *Message) ([]byte, error) {

	var fd int64
	var buf []byte

	if err := decodeWrite(m, &fd, &buf); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("write(%v, [%v]) -> %v\n", fd, len(buf), err)
		return result(err, nil)
	}

	n, err := file.Write(buf)
	if err != nil {
		log.Printf("write(%v, [%v]) -> %v\n", fd, len(buf), err)
		return result(err, nil)
	}

	log.Printf("write(%v, [%v]) -> %v\n", fd, len(buf), n)
	return result(nil, n)
}
