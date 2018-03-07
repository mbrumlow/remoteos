package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/mbrumlow/remoteos/pkg/remote"
)

func main() {

	// Make sure you have launched the remote systemcall provider.

	// Create temporary file for testing on local system.
	content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	log.Printf("Test File: %v\n", tmpfile.Name())

	// Connect to a remote syscall provider.
	r, err := remote.Connect(":7575")
	if err != nil {
		log.Fatal(err)
	}

	// Open the file over the remote interface.
	rFileA, err := r.Open(tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer rFileA.Close()

	// Create new file on the remote instance.
	rFileB, err := r.Create(tmpfile.Name() + ".copy")
	if err != nil {
		log.Fatal(err)
	}
	defer rFileB.Close()

	// Copy from one to the other.
	if n, err := io.Copy(rFileB, rFileA); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Copied %v bytes\n", n)
	}

	if ret, err := rFileA.Seek(0, 0); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Seek to %v on %v\n", ret, rFileA.Name())
	}

	if n, err := rFileB.WriteAt([]byte("test"), 1); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("WriteAt %v bytes\n", n)
	}

	buf := make([]byte, 4)
	if n, err := rFileB.ReadAt(buf, 1); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("ReadAt %v bytes -> [%v]\n", n, string(buf))
	}

	if err := rFileB.Sync(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Sync OK")
	}

}
