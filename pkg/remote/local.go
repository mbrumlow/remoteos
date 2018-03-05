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

	"github.com/mbrumlow/remoteos/pkg/proto"
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

	defer conn.Close()

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
		if buf, err := lh.read(m); err != nil {
			return err
		} else {
			return SendBuffer(conn, buf)
		}
	case SYS_WRITE:
		if buf, err := lh.write(m); err != nil {
			return err
		} else {
			return SendBuffer(conn, buf)
		}
	case SYS_OPEN:
		if buf, err := lh.open(m); err != nil {
			return err
		} else {
			return SendBuffer(conn, buf)
		}
	case SYS_CLOSE:
		return lh.sysClose(conn, m)
	case SYS_SEEK:
		if buf, err := lh.seek(m); err != nil {
			return err
		} else {
			return SendBuffer(conn, buf)
		}
	case SYS_PREAD64:
		return lh.sysReadAt(conn, m)
	case SYS_PWRITE64:
		return lh.sysWriteAt(conn, m)
	}

	// TODO return a unsupported error to caller.

	return nil
}

func (lh *LocalHost) sysReadAt(conn net.Conn, m *proto.Message) error {

	var fd int64
	var size int32
	var off int64

	log.Printf("sysReadAt decoding syscall buffer")
	if err := m.Decode(&fd, &size, &off); err != nil {
		log.Printf("sysRead syscall buffer ERR")
		return err
	}
	log.Printf("sysReadAt syscall buffer OK")

	file, ok := lh.LoadFile(fd)
	if !ok {
		log.Printf("os.ReadAt(%v, %v, %v) -> %v\n", fd, size, off, "Invalid argument")
		return SendError(conn, "Invalid argument")
	}

	buf := make([]byte, size)
	n, err := file.ReadAt(buf, off)
	if err != nil && err != io.EOF {
		log.Printf("os.Read(%v, %v, %v) -> %v\n", fd, size, off, errStr(err))
		return SendError(conn, errStr(err))
	}

	log.Printf("os.ReadAt(%v, %v, %v) -> ...\n", fd, size, off)
	return SendResult(conn, buf[:n])
}

func (lh *LocalHost) sysWriteAt(conn net.Conn, m *proto.Message) error {

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
		log.Printf("os.WriteAt(%v, %v, ...) -> %v\n", fd, len(buf), errStr(err))
		return SendError(conn, errStr(err))
	}

	log.Printf("os.WriteAt(%v, %v, ...) -> %v\n", fd, len(buf), n)
	return SendResult(conn, n)
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
