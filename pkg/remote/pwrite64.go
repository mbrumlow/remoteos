package remote

import (
	"bytes"
	"log"
)

func encodePwrite64(fd int64, b []byte, off int64) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.EncodeCall(SYS_PWRITE64, fd, b, off); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodePwrite64(m *Message, fd *int64, b *[]byte, off *int64) error {
	return m.Decode(fd, b, off)
}

func (rh *RemoteHost) pwrite64(fd int64, b []byte, off int64) (n int, err error) {

	// Encode
	call, err := encodePwrite64(fd, b, off)
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

func (lh *LocalHost) pwrite64(m *Message) ([]byte, error) {

	var fd int64
	var buf []byte
	var off int64

	if err := decodePwrite64(m, &fd, &buf, &off); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("pwrite64(%v, [%v], %v) -> %v\n", fd, len(buf), off, err)
		return result(err, nil)
	}

	n, err := file.WriteAt(buf, off)
	if err != nil {
		log.Printf("pwrite64(%v, [%v], %v) -> %v\n", fd, len(buf), off, err)
		return result(err, nil)
	}

	log.Printf("pwrite64(%v, [%v], %v) -> %v\n", fd, len(buf), off, n)
	return result(nil, n)
}
