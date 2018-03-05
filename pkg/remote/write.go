package remote

import (
	"bytes"
	"errors"

	"github.com/mbrumlow/remoteos/pkg/proto"
)

func encodeWrite(fd int64, b []byte) ([]byte, error) {
	m := proto.NewMessage(new(bytes.Buffer))
	if err := m.Encode(CMD_SYSCALL, SYS_WRITE, fd, b); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeWrite(m *proto.Message, fd *int64, b *[]byte) error {
	return m.Decode(fd, b)
}

func (rh *RemoteHost) write(fd int64, p []byte) (n int, err error) {

	// Encode
	call, err := encodeWrite(fd, p)
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

func (lh *LocalHost) write(m *proto.Message) ([]byte, error) {

	var fd int64
	var buf []byte

	if err := decodeWrite(m, &fd, &buf); err != nil {
		return nil, err
	}

	file, ok := lh.LoadFile(fd)
	if !ok {
		return result(errors.New("Invalid arguemtn"), nil)
	}

	n, err := file.Write(buf)
	if err != nil {
		return result(err, nil)
	}

	return result(nil, n)
}
