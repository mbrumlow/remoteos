package remote

import (
	"bytes"
	"log"
)

func encodeSync(fd int64) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.EncodeCall(SYS_SYNC, fd); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeSync(m *Message, fd *int64) error {
	return m.Decode(fd)
}

func (rh *RemoteHost) sync(fd int64) error {

	// Encode
	call, err := encodeSync(fd)
	if err != nil {
		return err
	}

	// Send
	_, err = rh.sendCall2(call)
	if err != nil {
		return err
	}

	return nil
}

func (lh *LocalHost) sync(m *Message) ([]byte, error) {

	var fd int64

	if err := decodeSync(m, &fd); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("sync(%v) -> %v\n", err)
		return result(err, nil)
	}

	if err := file.Sync(); err != nil {
		log.Printf("sync(%v) -> %v\n", err)
		return result(err, nil)
	}

	log.Printf("sync(%v) -> OK\n", fd)
	return result(nil)
}
