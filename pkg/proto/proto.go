package proto

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

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
		case uintptr:
			x := uint64(v)
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
		case *uintptr:
			var x uint64
			bRead(&x)
			*v = uintptr(x)
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
