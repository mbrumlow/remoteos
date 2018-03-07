package remote

import (
	"bytes"
	"log"
)

func encodeClose(fd int64) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.EncodeCall(SYS_CLOSE, fd); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeClose(m *Message, fd *int64) error {
	return m.Decode(fd)
}

func (rh *RemoteHost) close(fd int64) error {

	// Encode
	call, err := encodeClose(fd)
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

func (lh *LocalHost) close(m *Message) ([]byte, error) {

	var fd int64

	if err := decodeClose(m, &fd); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("close(%v) -> %v\n", fd, err)
		return result(err, nil)
	}

	defer func() {
		lh.mu.Lock()
		defer lh.mu.Unlock()
		delete(lh.FDs, fd)
	}()

	if err := file.Close(); err != nil {
		log.Printf("close(%v) -> %v\n", fd, err)
		return result(err, nil)
	}

	log.Printf("close(%v) -> OK\n", fd)
	return result(nil)
}
