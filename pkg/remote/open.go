package remote

import (
	"bytes"
	"os"

	"github.com/mbrumlow/remoteos/pkg/proto"
)

func encodeOpen(name string, flag int, perm os.FileMode) ([]byte, error) {
	m := proto.NewMessage(new(bytes.Buffer))
	if err := m.Encode(CMD_SYSCALL, SYS_OPEN, name, flag, perm); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeOpen(m *proto.Message, name *string, flag *int, perm *os.FileMode) error {
	return m.Decode(name, flag, perm)
}

func (rh *RemoteHost) open(name string, flag int, perm os.FileMode) (fd int64, err error) {

	// Encode
	call, err := encodeOpen(name, flag, perm)
	if err != nil {
		return 0, err
	}

	// Send
	m, err := rh.sendCall2(call)
	if err != nil {
		return 0, err
	}

	// Handle reply
	if err := m.Decode(&fd); err != nil {
		return 0, err
	}

	return fd, nil
}

func (lh *LocalHost) open(m *proto.Message) ([]byte, error) {

	var name string
	var flag int
	var perm os.FileMode

	if err := decodeOpen(m, &name, &flag, &perm); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return result(err, nil)
	}

	lh.StoreFile(file)

	return result(nil, file.Fd())
}
