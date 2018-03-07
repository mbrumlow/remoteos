package remote

import (
	"bytes"
	"fmt"

	"github.com/mbrumlow/remoteos/pkg/proto"
)

type Message struct {
	bb *bytes.Buffer
}

func NewMessage(bb *bytes.Buffer) *Message {
	return &Message{bb: bb}
}

func (m *Message) Bytes() []byte {
	return m.bb.Bytes()
}

func (m *Message) Decode(a ...interface{}) error {
	return proto.Decode(m.bb, a...)
}

func (m *Message) Encode(a ...interface{}) error {
	return proto.Encode(m.bb, a...)
}

func (m *Message) EncodeCall(call RemoteCall, a ...interface{}) error {
	if err := proto.Encode(m.bb, call); err != nil {
		return err
	}
	if err := proto.Encode(m.bb, a...); err != nil {
		return err
	}
	return nil
}

func (m *Message) EncodeError(s string) error {
	return proto.Encode(m.bb, int32(1), s)
}

func (m *Message) DecodeError() error {
	var s string
	if err := proto.Decode(m.bb, &s); err != nil {
		return err
	}
	return fmt.Errorf("%v", s)
}

func (m *Message) EncodeResult(a ...interface{}) error {
	if err := proto.Encode(m.bb, int32(0)); err != nil {
		return err
	}
	if err := proto.Encode(m.bb, a...); err != nil {
		return err
	}

	return nil
}
