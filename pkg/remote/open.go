package remote

import (
	"bytes"
	"log"
	"os"
)

func encodeOpen(name string, flag int, perm os.FileMode) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.Encode(SYS_OPEN, name, flag, perm); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeOpen(m *Message, name *string, flag *int, perm *os.FileMode) error {
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

func (lh *LocalHost) open(m *Message) ([]byte, error) {

	var name string
	var flag int
	var perm os.FileMode

	if err := decodeOpen(m, &name, &flag, &perm); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		log.Printf("open(%v, %v, %v) -> %v\n", name, flag, perm, err)
		return result(err, nil)
	}

	lh.StoreFile(file)

	log.Printf("open(%v, %v, %v) -> %v\n", name, flag, perm, file.Fd())
	return result(nil, file.Fd())
}
