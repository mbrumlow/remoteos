package remote

import (
	"bytes"
	"testing"

	"github.com/mbrumlow/remoteos/pkg/proto"
)

func TestReadEncodeDecode(t *testing.T) {

	expectFD := int64(123)
	expectSize := int(1000)

	call, err := encodeRead(expectFD, expectSize)
	if err != nil {
		t.Errorf("Encode failed: %v", err)
	}

	m := proto.NewMessage(bytes.NewBuffer(call))

	var cmd uint32
	var syscall uint32
	if err := m.Decode(&cmd, &syscall); err != nil {
		t.Errorf("Decode call header failed: %v", err)
	}

	if cmd != CMD_SYSCALL {
		t.Errorf("Expected cmd %v, but was %v", CMD_SYSCALL, cmd)
	}

	if syscall != 0 {
		t.Errorf("Expected cmd %v, but was %v", SYS_READ, syscall)
	}

	var fd int64
	var size int
	if err := decodeRead(m, &fd, &size); err != nil {
		t.Errorf("Decode failed: %v", err)
	}

	if fd != expectFD {
		t.Errorf("Expected fd %v, but was %v", expectFD, fd)
	}
}
