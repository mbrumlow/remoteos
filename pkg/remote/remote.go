package remote

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/mbrumlow/remoteos/pkg/proto"
)

const (
	CMD_SYSCALL = uint32(0)

	SYS_READ     = uint32(0)
	SYS_WRITE    = uint32(1)
	SYS_OPEN     = uint32(2)
	SYS_CLOSE    = uint32(3)
	SYS_SEEK     = uint32(8)
	SYS_PREAD64  = uint32(17)
	SYS_PWRITE64 = uint32(18)
)

type RemoteHost struct {
	conn net.Conn
}

func Connect(host string) (*RemoteHost, error) {

	// TODO add auth and encryption.

	conn, err := net.Dial("tcp", host)
	return &RemoteHost{conn: conn}, err
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

func SendError(conn net.Conn, s string) error {
	ret := new(bytes.Buffer)
	proto.EncodeError(ret, s)
	return SendBuffer(conn, ret.Bytes())
}

func SendResult(conn net.Conn, a ...interface{}) error {
	ret := new(bytes.Buffer)
	proto.EncodeResult(ret, a...)
	return SendBuffer(conn, ret.Bytes())
}

func result(e error, a ...interface{}) ([]byte, error) {

	ret := new(bytes.Buffer)

	if e != nil {
		if err := proto.EncodeError(ret, errStr(e)); err != nil {
			return nil, err
		}
		return ret.Bytes(), nil
	}

	if err := proto.EncodeResult(ret, a...); err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
}

func reciveBuffer(conn net.Conn) (*bytes.Buffer, error) {
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

func (rh *RemoteHost) sendCall2(b []byte) (*proto.Message, error) {

	if err := SendBuffer(rh.conn, b); err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	bb, err := reciveBuffer(rh.conn)
	if err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	m := proto.NewMessage(bb)

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

func (rh *RemoteHost) sendCall(call uint32, a ...interface{}) (*proto.Message, error) {

	request := new(bytes.Buffer)
	proto.EncodeCall(request, CMD_SYSCALL, call, a...)

	if err := SendBuffer(rh.conn, request.Bytes()); err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	bb, err := reciveBuffer(rh.conn)
	if err != nil {
		// TODO: figure out what to do.
		// TODO: Mabye sessions for reconnect?
		return nil, err
	}

	m := proto.NewMessage(bb)
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
