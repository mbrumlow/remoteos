package remote

import (
	"bytes"
	"io"
	"log"
)

func encodeRead(fd int64, size int) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.Encode(SYS_READ, fd, size); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeRead(m *Message, fd *int64, size *int) error {
	return m.Decode(fd, size)
}

func (rh *RemoteHost) read(fd int64, p []byte) (n int, err error) {

	// Encode
	call, err := encodeRead(fd, len(p))
	if err != nil {
		return 0, err
	}

	// Send
	m, err := rh.sendCall2(call)
	if err != nil {
		return 0, err
	}

	// Handle reply
	var buf []byte
	if err := m.Decode(&buf); err != nil {
		return 0, err
	}

	if len(buf) == 0 {
		return 0, io.EOF
	}

	copy(p, buf)
	return len(buf), nil
}

func (lh *LocalHost) read(m *Message) ([]byte, error) {

	var fd int64
	var size int

	if err := decodeRead(m, &fd, &size); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("read(%v, %v) -> %v\n", fd, size, err)
		return result(err, nil)
	}

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("read(%v, %v) -> %v\n", fd, size, err)
		return result(err, nil)
	}

	log.Printf("read(%v, %v) -> [%v]\n", fd, size, n)
	return result(nil, buf[:n])
}
