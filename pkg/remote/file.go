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
	n, err = f.rh.read(f.fd, p)
	if err != nil && err != io.EOF {
		return 0, &os.PathError{"read", f.name, err}
	}
	return
}

func (f *File) Write(b []byte) (n int, err error) {
	n, err = f.rh.write(f.fd, b)
	if err != nil {
		return 0, &os.PathError{"write", f.name, err}
	}
	return
}

func (f *File) pread64(p []byte, off int64) (n int, err error) {
	n, err = f.rh.pread64(f.fd, p, off)
	if err != nil && err != io.EOF {
		return 0, &os.PathError{"read", f.name, err}
	}
	return
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	return f.pread64(b, off)
}

func (f *File) pwrite64(b []byte, off int64) (n int, err error) {
	n, err = f.rh.pwrite64(f.fd, b, off)
	if err != nil {
		return 0, &os.PathError{"write", f.name, err}
	}
	return
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	return f.pwrite64(b, off)
}

func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	ret, err = f.rh.seek(f.fd, offset, whence)
	if err != nil {
		return 0, &os.PathError{"seek", f.name, err}
	}
	return
}

func (f *File) Sync() error {
	err := f.rh.sync(f.fd)
	if err != nil {
		return &os.PathError{"sync", f.name, err}
	}
	return nil
}

func (f *File) Close() error {
	err := f.rh.close(f.fd)
	if err != nil {
		return &os.PathError{"close", f.name, err}
	}
	return nil
}
