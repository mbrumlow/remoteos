package remote

import (
	"io"
	"os"
)

type File struct {
	fd   int64
	name string
	rh   *RemoteHost
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Read(p []byte) (n int, err error) {
	m, err := f.rh.sendCall(SYS_READ, f.fd, int32(len(p)))
	if err != nil {
		return 0, &os.PathError{"read", f.name, err}
	}

	var buf []byte
	if err := m.Decode(&buf); err != nil {
		return 0, &os.PathError{"read", f.name, err}
	}

	if len(buf) == 0 {
		return 0, io.EOF
	}

	copy(p, buf)

	return len(buf), nil
}

func (f *File) pread64(p []byte, off int64) (n int, err error) {

	m, err := f.rh.sendCall(SYS_PREAD64, f.fd, off, int32(len(p)))
	if err != nil {
		return 0, &os.PathError{"read", f.name, err}
	}

	var buf []byte
	if err := m.Decode(&buf); err != nil {
		return 0, &os.PathError{"read", f.name, err}
	}

	if len(buf) == 0 {
		return 0, io.EOF
	}

	copy(p, buf)

	return len(buf), nil
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	return f.pread64(b, off)
}

func (f *File) Write(b []byte) (n int, err error) {
	m, err := f.rh.sendCall(SYS_WRITE, f.fd, b)
	if err != nil {
		return 0, &os.PathError{"write", f.name, err}
	}

	if err := m.Decode(&n); err != nil {
		return 0, &os.PathError{"write", f.name, err}
	}

	return
}

func (f *File) pwrite64(b []byte, off int64) (n int, err error) {

	m, err := f.rh.sendCall(SYS_PWRITE64, f.fd, off, int32(len(b)), b)
	if err != nil {
		return 0, &os.PathError{"write", f.name, err}
	}

	if err := m.Decode(&n); err != nil {
		return 0, &os.PathError{"write", f.name, err}
	}

	return
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	return f.pwrite64(b, off)
}

func (f *File) Close() error {
	_, err := f.rh.sendCall(SYS_CLOSE, f.fd)
	if err != nil {
		return &os.PathError{"close", f.name, err}
	}
	return err
}

func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	m, err := f.rh.sendCall(SYS_SEEK, f.fd, offset, whence)
	if err != nil {
		return 0, &os.PathError{"seek", f.name, err}
	}

	if err := m.Decode(&ret); err != nil {
		return 0, &os.PathError{"seek", f.name, err}
	}

	return
}
