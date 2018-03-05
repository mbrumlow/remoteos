package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
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
	return Decode(m.bb, a...)
}

func (m *Message) Encode(a ...interface{}) error {
	return Encode(m.bb, a...)
}

func (m *Message) DecodeError() error {
	return DecodeError(m.bb)
}

func Encode(bb *bytes.Buffer, a ...interface{}) (err error) {

	bWrite := func(x interface{}) {
		if err != nil {
			return
		}
		err = binary.Write(bb, binary.BigEndian, x)
	}

	for i := range a {

		if err != nil {
			return
		}

		switch v := a[i].(type) {
		case int:
			x := int32(v)
			bWrite(x)
		case os.FileMode:
			x := int32(v)
			bWrite(x)
		case string:
			bWrite(uint32(len(v)))
			bb.WriteString(v)
		case []byte:
			bWrite(uint32(len(v)))
			bb.Write(v)
		default:
			bWrite(v)
		}
	}

	return err

}

func Decode(bb *bytes.Buffer, a ...interface{}) (err error) {

	bRead := func(x interface{}) {
		if err != nil {
			return
		}
		err = binary.Read(bb, binary.BigEndian, x)
	}

	for i := range a {

		if err != nil {
			return
		}

		switch v := a[i].(type) {
		case *int:
			var x int32
			bRead(&x)
			*v = int(x)
		case *os.FileMode:
			var x int32
			bRead(&x)
			*v = os.FileMode(x)
		case *string:
			var size uint32
			bRead(&size)
			if bb.Len() < int(size) {
				return io.EOF
			}
			*v = string(bb.Next(int(size)))
		case *[]byte:
			var size uint32
			bRead(&size)
			if bb.Len() < int(size) {
				return io.EOF
			}
			*v = bb.Next(int(size))
		default:
			bRead(v)
		}
	}

	return err
}

func EncodeCall(bb *bytes.Buffer, cmd, call uint32, a ...interface{}) {
	Encode(bb, cmd)
	Encode(bb, call)
	Encode(bb, a...)
}

func EncodeError(bb *bytes.Buffer, s string) error {
	return Encode(bb, int32(1), s)
}

func EncodeResult(bb *bytes.Buffer, a ...interface{}) error {
	if err := Encode(bb, int32(0)); err != nil {
		return err
	}
	if err := Encode(bb, a...); err != nil {
		return err
	}

	return nil
}

func DecodeError(bb *bytes.Buffer) error {
	var s string
	if err := Decode(bb, &s); err != nil {
		return err
	}
	return fmt.Errorf("%v", s)
}
