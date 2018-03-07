package remote

import (
	"bytes"
	"io"
	"log"
)

func encodePread64(fd int64, p []byte, off int64) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.EncodeCall(SYS_PREAD64, fd, len(p), off); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodePread64(m *Message, fd *int64, size *int, off *int64) error {
	return m.Decode(fd, size, off)
}

func (rh *RemoteHost) pread64(fd int64, p []byte, off int64) (n int, err error) {

	// Encode
	call, err := encodePread64(fd, p, off)
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

func (lh *LocalHost) pread64(m *Message) ([]byte, error) {

	var fd int64
	var size int
	var off int64

	if err := decodePread64(m, &fd, &size, &off); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("pread64(%v, %v, %v) -> %v\n", fd, size, off, err)
		return result(err, nil)
	}

	buf := make([]byte, size)
	n, err := file.ReadAt(buf, off)
	if err != nil && err != io.EOF {
		log.Printf("pread64(%v, %v, %v) -> %v\n", fd, size, off, err)
		return result(err, nil)
	}

	log.Printf("pread64(%v, %v, %v) -> [%v]\n", fd, size, off, n)
	return result(nil, buf[:n])

}
