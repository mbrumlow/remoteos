package remote

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type LocalHost struct {
	mu  sync.Mutex
	FDs map[int64]*os.File

	sendMutex sync.Mutex
}

func NewLocalHost(network string) *LocalHost {
	return &LocalHost{
		FDs: make(map[int64]*os.File),
	}
}

func (lh *LocalHost) Run() error {

	ln, err := net.Listen("tcp", ":7575")
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		go lh.handleConnection(conn)
	}
}

func (lh *LocalHost) handleConnection(conn net.Conn) {

	log.Printf("Handling connection")

	defer conn.Close()

	for {

		var len uint32
		if err := binary.Read(conn, binary.BigEndian, &len); err != nil {
			log.Println(err)
			return
		}
		// TODO: make sure len is sane.

		var id uint64
		if err := binary.Read(conn, binary.BigEndian, &id); err != nil {
			log.Println(err)
			return
		}

		// Read message in full.
		buf := make([]byte, len)
		if _, err := io.ReadFull(conn, buf); err != nil {
			log.Println(err)
			return
		}

		bb := bytes.NewBuffer(buf)

		go func() {
			if err := lh.handleSysCall(conn, id, NewMessage(bb)); err != nil {
				log.Printf("Fatal error: %v", err)
			}
		}()
	}
}

func (lh *LocalHost) sendBuffer(conn net.Conn, id uint64, b []byte) error {
	lh.sendMutex.Lock()
	defer lh.sendMutex.Unlock()
	return sendBuffer2(conn, id, b)
}

func (lh *LocalHost) handleSysCall(conn net.Conn, id uint64, m *Message) (err error) {

	var syscall RemoteCall
	if err := m.Decode(&syscall); err != nil {
		return err
	}

	buf, err := lh.call(conn, syscall, m)
	if err != nil {
		return err
	}

	err = lh.sendBuffer(conn, id, buf)
	return err
}

func (lh *LocalHost) call(conn net.Conn, syscall RemoteCall, m *Message) ([]byte, error) {

	switch syscall {
	case SYS_READ:
		return lh.read(m)
	case SYS_WRITE:
		return lh.write(m)
	case SYS_OPEN:
		return lh.open(m)
	case SYS_CLOSE:
		return lh.close(m)
	case SYS_SEEK:
		return lh.seek(m)
	case SYS_PREAD64:
		return lh.pread64(m)
	case SYS_PWRITE64:
		return lh.pwrite64(m)
	case SYS_SYNC:
		return lh.sync(m)
	}

	// TODO generate buffer error to send back to host, lets not make this
	// a proto error.
	return nil, fmt.Errorf("call not supported.")
}

func (lh *LocalHost) StoreFile(file *os.File) {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	lh.FDs[int64(file.Fd())] = file
}

func (lh *LocalHost) LoadFile(fd int64) (file *os.File, err error) {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	file, ok := lh.FDs[fd]
	if !ok {
		return nil, errors.New("Invalid argument")
	}
	return file, nil
}

func errStr(err error) string {
	if e, ok := err.(*os.PathError); ok {
		return fmt.Sprintf("%v", e.Err)
	}
	return fmt.Sprintf("%v", err)
}
