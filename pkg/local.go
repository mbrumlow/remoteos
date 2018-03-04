package remote

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/mbrumlow/remote/pkg/proto"
)

type LocalHost struct {
	mu  sync.Mutex
	FDs map[int64]*os.File
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

	for {

		var len uint32
		if err := binary.Read(conn, binary.BigEndian, &len); err != nil {
			log.Println(err)
			return
		}

		// TODO: make sure len is sane.
		log.Printf("new msg: %v\n", len)

		// Read message in full.
		buf := make([]byte, len)
		if n, err := io.ReadFull(conn, buf); err != nil {
			log.Println(err)
			return
		} else {
			log.Printf("read %v bytes\n", n)
		}

		bb := bytes.NewBuffer(buf)

		var cmd uint32
		binary.Read(bb, binary.BigEndian, &cmd)

		log.Printf("cmd: %v\n", cmd)
		if err := lh.handleSysCall(conn, proto.NewMessage(bb)); err != nil {
			log.Printf("Fatal error: %v", err)
			return
		}

	}
}

func (lh *LocalHost) handleSysCall(conn net.Conn, m *proto.Message) error {

	var syscall uint32
	if err := m.Decode(&syscall); err != nil {
		return err
	}

	log.Printf("syscall: %v", syscall)

	switch syscall {
	case SYS_READ:
		return lh.sysRead(conn, m)
	case SYS_WRITE:
		return lh.sysWrite(conn, m)
	case SYS_OPEN:
		return lh.sysOpen(conn, m)
	case SYS_CLOSE:
		return lh.sysClose(conn, m)
	case SYS_SEEK:
		return lh.sysSeek(conn, m)
	}

	// TODO return a unsupported error to caller.

	return nil
}

func (lh *LocalHost) sysRead(conn net.Conn, m *proto.Message) error {

	var fd int64
	var size int32

	log.Printf("sysRead decoding syscall buffer")
	if err := m.Decode(&fd, &size); err != nil {
		log.Printf("sysRead syscall buffer ERR")
		return err
	}
	log.Printf("sysRead syscall buffer OK")

	file, ok := lh.LoadFile(fd)
	if !ok {
		log.Printf("os.Read(%v, %v) -> %v\n", fd, size, "Invalid argument")
		return SendError(conn, "Invalid argument")
	}

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("os.Read(%v, %v) -> %v\n", fd, size, errStr(err))
		return SendError(conn, errStr(err))
	}

	log.Printf("os.Read(%v, %v) -> ...\n", fd, size)
	return SendResult(conn, buf[:n])
}

func (lh *LocalHost) sysWrite(conn net.Conn, m *proto.Message) error {

	var fd int64
	var buf []byte

	log.Printf("sysWrite decoding syscall buffer")
	if err := m.Decode(&fd, &buf); err != nil {
		log.Printf("sysWrite syscall buffer ERR")
		return err
	}
	log.Printf("sysWrite syscall buffer OK")

	file, ok := lh.LoadFile(fd)
	if !ok {
		log.Printf("os.Write(%v, %v, ...) -> %v\n", fd, len(buf), "Invalid argument")
		return SendError(conn, "Invalid argument")
	}

	n, err := file.Write(buf)
	if err != nil && err != io.EOF {
		log.Printf("os.Write(%v, %v, ...) -> %v\n", fd, len(buf), errStr(err))
		return SendError(conn, errStr(err))
	}

	log.Printf("os.Write(%v, %v, ...) -> %v\n", fd, len(buf), n)
	return SendResult(conn, int32(n))
}

func (lh *LocalHost) hosWriteAt(conn net.Conn, m proto.Message) error {

	var fd int64
	var offset int64
	var buf []byte

	log.Printf("sysWrite decoding syscall buffer")
	if err := m.Decode(fd, offset, buf); err != nil {
		return err
	}
	log.Printf("sysWrite syscall buffer OK")

	file, ok := lh.LoadFile(fd)
	if !ok {
		log.Printf("os.WriteAt(%v, %v, ...) -> %v\n", fd, len(buf), "Invalid argument")
		return SendError(conn, "Invalid argument")
	}

	n, err := file.WriteAt(buf, offset)
	if err != nil && err != io.EOF {
		log.Printf("os.Write(%v, %v, ...) -> %v\n", fd, len(buf), errStr(err))
		return SendError(conn, errStr(err))
	}

	log.Printf("os.Write(%v, %v, ...) -> %v\n", fd, len(buf), n)
	return SendResult(conn, int32(n))
}

func (lh *LocalHost) sysOpen(conn net.Conn, m *proto.Message) error {

	var name string
	var flag int
	var perm os.FileMode

	log.Printf("sysOpen decoding syscall buffer")
	if err := m.Decode(&name, &flag, &perm); err != nil {
		log.Printf("sysOpen syscall buffer ERR")
		return err
	}
	log.Printf("sysOpen syscall buffer OK")

	file, err := os.OpenFile(name, int(flag), os.FileMode(perm))
	if err != nil {
		log.Printf("os.Open(%v) -> %v\n", name, errStr(err))
		return SendError(conn, errStr(err))
	}

	lh.StoreFile(file)

	log.Printf("os.Open(%v) -> %v\n", name, file.Fd())
	return SendResult(conn, int64(file.Fd()))
}

func (lh *LocalHost) sysClose(conn net.Conn, m *proto.Message) error {

	var fd int64

	log.Printf("sysClose decoding syscall buffer")
	if err := m.Decode(&fd); err != nil {
		log.Printf("sysClose syscall buffer ERR")
		return err
	}
	log.Printf("sysClose syscall buffer OK")

	file, ok := lh.LoadFile(fd)
	if !ok {
		log.Printf("os.Close(%v) -> %v\n", fd, "Invalid argument")
		return SendError(conn, "Invalid argument")
	}

	defer func() {
		lh.mu.Lock()
		defer lh.mu.Unlock()
		delete(lh.FDs, fd)
	}()

	if err := file.Close(); err != nil {
		log.Printf("os.Close(%v) -> %v\n", fd, err)
		return SendError(conn, fmt.Sprintf("%v", err))
	}

	log.Printf("os.Close(%v) -> 0\n", fd)
	return SendResult(conn)
}

func (lh *LocalHost) sysSeek(conn net.Conn, m *proto.Message) error {

	var fd int64
	var offset int64
	var whence int

	log.Printf("sysSeek decoding syscall buffer")
	if err := m.Decode(&fd, &offset, &whence); err != nil {
		log.Printf("sysSeek syscall buffer ERR")
		return err
	}
	log.Printf("sysSeek syscall buffer OK")

	file, ok := lh.LoadFile(fd)
	if !ok {
		log.Printf("os.Seek(%v, %v, %v) -> %v\n", fd, offset, whence, "Invalid argument")
		return SendError(conn, "Invalid argument")
	}

	n, err := file.Seek(offset, whence)
	if err != nil {
		log.Printf("os.Seeek(%v, %v, %v) -> %v\n", fd, offset, whence, errStr(err))
		return SendError(conn, errStr(err))
	}

	log.Printf("os.Seeek(%v, %v, %v) -> %v\n", fd, offset, whence, n)
	return SendResult(conn, n)
}

func (lh *LocalHost) StoreFile(file *os.File) {

	lh.mu.Lock()
	defer lh.mu.Unlock()
	lh.FDs[int64(file.Fd())] = file

}

func (lh *LocalHost) LoadFile(fd int64) (file *os.File, ok bool) {

	lh.mu.Lock()
	defer lh.mu.Unlock()
	file, ok = lh.FDs[fd]
	return
}

func errStr(err error) string {
	if e, ok := err.(*os.PathError); ok {
		return fmt.Sprintf("%v", e.Err)
	}
	return fmt.Sprintf("%v", err)
}
