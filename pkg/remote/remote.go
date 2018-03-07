package remote

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

type RemoteCall uint32

const (
	SYS_READ     = RemoteCall(0)
	SYS_WRITE    = RemoteCall(1)
	SYS_OPEN     = RemoteCall(2)
	SYS_CLOSE    = RemoteCall(3)
	SYS_SEEK     = RemoteCall(8)
	SYS_PREAD64  = RemoteCall(17)
	SYS_PWRITE64 = RemoteCall(18)
	SYS_SYNC     = RemoteCall(162)
)

func (rc RemoteCall) String() string {
	switch rc {
	case 0:
		return "read"
	case 1:
		return "write"
	case 2:
		return "open"
	case 3:
		return "close"
	case 8:
		return "seek"
	case 17:
		return "pread64"
	case 18:
		return "pwrite64"
	case 162:
		return "sync"
	default:
		return fmt.Sprintf("%d", rc)
	}
}

type RemoteHost struct {
	conn      net.Conn
	callCount uint64

	sendMutex sync.Mutex

	receiveMutex sync.Mutex
	bufferMap    map[uint64]*bytes.Buffer
}

func Connect(host string) (*RemoteHost, error) {

	// TODO add auth and encryption.

	conn, err := net.Dial("tcp", host)
	return &RemoteHost{conn: conn,
		bufferMap: make(map[uint64]*bytes.Buffer),
	}, err
}

func SendBuffer(conn net.Conn, b []byte) error {
	size := uint32(len(b))
	if err := binary.Write(conn, binary.BigEndian, size); err != nil {
		return err
	}

	// TODO consider compression.

	if _, err := conn.Write(b); err != nil {
		return err
	}

	return nil
}

func sendBuffer2(conn io.Writer, id uint64, b []byte) error {
	size := uint32(len(b))
	if err := binary.Write(conn, binary.BigEndian, size); err != nil {
		return err
	}

	if err := binary.Write(conn, binary.BigEndian, id); err != nil {
		return err
	}

	// TODO consider compression.

	if _, err := conn.Write(b); err != nil {
		return err
	}

	return nil
}

func (rh *RemoteHost) sendBuffer(id uint64, b []byte) error {
	rh.sendMutex.Lock()
	defer rh.sendMutex.Unlock()
	return sendBuffer2(rh.conn, id, b)
}

func (rh *RemoteHost) receiveBuffer(conn net.Conn, id uint64) (*bytes.Buffer, error) {

	defer rh.receiveMutex.Unlock()

	for {
		rh.receiveMutex.Lock()

		if bb, ok := rh.bufferMap[id]; ok {
			return bb, nil
		}

		var size uint32
		if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
			return nil, err
		}

		var bid uint64
		if err := binary.Read(conn, binary.BigEndian, &bid); err != nil {
			return nil, err
		}

		buf := make([]byte, size)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, err
		}

		if id == bid {
			return bytes.NewBuffer(buf), nil
		} else {
			rh.bufferMap[bid] = bytes.NewBuffer(buf)
		}

		rh.receiveMutex.Unlock()
	}

	return nil, fmt.Errorf("Failed to receive buffer")
}

/*
func SendError(conn net.Conn, s string) error {
	ret := new(bytes.Buffer)
	EncodeError(ret, s)
	return SendBuffer(conn, ret.Bytes())
}
*/

/*
func SendResult(conn net.Conn, a ...interface{}) error {
	ret := new(bytes.Buffer)
	EncodeResult(ret, a...)
	return SendBuffer(conn, ret.Bytes())
}
*/

func result(e error, a ...interface{}) ([]byte, error) {

	m := NewMessage(new(bytes.Buffer))

	if e != nil {
		if err := m.EncodeError(errStr(e)); err != nil {
			return nil, err
		}
		return m.Bytes(), nil
	}

	if err := m.EncodeResult(a...); err != nil {
		return nil, err
	}

	return m.Bytes(), nil
}

func receiveBuffer(conn net.Conn) (*bytes.Buffer, error) {
	var size uint32
	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return nil, err
	}

	buf := make([]byte, size)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	return bytes.NewBuffer(buf), nil
}

func (rh *RemoteHost) sendCall2(b []byte) (*Message, error) {

	id := atomic.AddUint64(&rh.callCount, 1)

	if err := rh.sendBuffer(id, b); err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	bb, err := rh.receiveBuffer(rh.conn, id)
	if err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	m := NewMessage(bb)

	var ret int32
	if err := m.Decode(&ret); err != nil {
		return m, err
	}

	switch ret {
	case 0:
		return m, nil
	case 1:
		return m, m.DecodeError()
	case 2:
		return m, fmt.Errorf("unsupported")
	}

	return m, fmt.Errorf("unknown")
}

func (rh *RemoteHost) sendCall(call RemoteCall, a ...interface{}) (*Message, error) {

	req := NewMessage(new(bytes.Buffer))
	req.EncodeCall(call, a...)

	if err := SendBuffer(rh.conn, req.Bytes()); err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	bb, err := receiveBuffer(rh.conn)
	if err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	m := NewMessage(bb)
	var ret int32

	if err := m.Decode(&ret); err != nil {
		return nil, err
	}

	switch ret {
	case 0:
		return m, nil
	case 1:
		return m, m.DecodeError()
	case 2:
		return m, fmt.Errorf("unsupported")
	}

	return m, fmt.Errorf("unknown")
}

func (rh *RemoteHost) OpenFile(name string, flag int, perm os.FileMode) (*File, error) {
	fd, err := rh.open(name, flag, perm)
	if err != nil {
		return nil, &os.PathError{"open", name, err}
	}
	return &File{fd: fd, name: name, rh: rh}, nil
}

func (rh *RemoteHost) Open(name string) (*File, error) {
	return rh.OpenFile(name, os.O_RDONLY, 0)
}

func (rh *RemoteHost) Create(name string) (*File, error) {
	return rh.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}
