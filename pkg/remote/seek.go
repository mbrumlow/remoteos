package remote

import (
	"bytes"
	"log"
)

func encodeSeek(fd, offset int64, whence int) ([]byte, error) {
	m := NewMessage(new(bytes.Buffer))
	if err := m.EncodeCall(SYS_SEEK, fd, offset, whence); err != nil {
		return nil, err
	}
	return m.Bytes(), nil
}

func decodeSeek(m *Message, fd, offset *int64, whence *int) error {
	return m.Decode(fd, offset, whence)
}

func (rh *RemoteHost) seek(fd, offset int64, whence int) (ret int64, err error) {

	// Encode
	call, err := encodeSeek(fd, offset, whence)
	if err != nil {
		return 0, err
	}

	// Send
	m, err := rh.sendCall2(call)
	if err != nil {
		return 0, err
	}

	// Handle reply
	if err := m.Decode(&ret); err != nil {
		return 0, err
	}

	return ret, nil
}

func (lh *LocalHost) seek(m *Message) ([]byte, error) {

	var fd int64
	var offset int64
	var whence int

	if err := decodeSeek(m, &fd, &offset, &whence); err != nil {
		return nil, err
	}

	file, err := lh.LoadFile(fd)
	if err != nil {
		log.Printf("seek(%v, %v, %v) -> %v\n", fd, offset, whence, err)
		return result(err, nil)
	}

	ret, err := file.Seek(offset, whence)
	if err != nil {
		log.Printf("seek(%v, %v, %v) -> %v\n", fd, offset, whence, err)
		return result(err, nil)
	}

	log.Printf("seek(%v, %v, %v) -> %v\n", fd, offset, whence, ret)
	return result(nil, ret)
}
