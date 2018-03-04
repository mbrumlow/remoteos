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

	// Create new file on the remote instance.
	rFileB, err := r.Create(tmpfile.Name() + ".copy")
	if err != nil {
		log.Fatal(err)
	}

	// Copy from one to the other.
	n, err := io.Copy(rFileB, rFileA)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Copied %v bytes\n", n)

}
