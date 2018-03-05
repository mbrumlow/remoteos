package remote

import (
	"bytes"
	"errors"
	"io"

	"github.com/mbrumlow/remoteos/pkg/proto"
)

func encodeRead(fd int64, size int) ([]byte, error) {
	m := proto.NewMessage(new(bytes.Buffer))
	if err := m.Encode(CMD_SYSCALL, SYS_READ, fd, size); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeRead(m *proto.Message, fd *int64, size *int) error {
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

func (lh *LocalHost) read(m *proto.Message) ([]byte, error) {

	var fd int64
	var size int

	if err := decodeRead(m, &fd, &size); err != nil {
		return nil, err
	}

	file, ok := lh.LoadFile(fd)
	if !ok {
		return result(errors.New("Invalid arguemtn"), nil)
	}

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return result(err, nil)
	}

	return result(nil, buf[:n])
}
