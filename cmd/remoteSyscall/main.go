package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/mbrumlow/remoteos/pkg/remote"
)

var child = flag.Bool("child", false, "Runs in child mode")

func main() {

	flag.Parse()

	log.Printf("PATH: %v\n", os.Args[0])

	if *child {
		childProcess()
	} else {
		parrentProcess()
	}

}

func parrentProcess() {

	ln, err := net.Listen("tcp", ":7575")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConnection(conn)
	}

}

func childProcess() {

	conIn := os.NewFile(3, "connIn")
	conOut := os.NewFile(4, "conOut")

	// TODO handle errors -- check for null on handles.

	lh := remote.NewLocalHost()
	lh.Run(conIn, conOut)
}

func handleConnection(conn net.Conn) {

	defer conn.Close()
	cR, pW, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	defer cR.Close()
	defer pW.Close()

	pR, cW, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	defer pR.Close()
	defer cW.Close()

	cmd := exec.Command(os.Args[0], "-child")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{
		cR,
		cW,
	}

	log.Printf("Got new connection, forking...\n")

	// TODO handle errors.
	cmd.Start()

	go func() {
		io.Copy(pW, conn)
		pW.Close()
	}()

	go func() {
		io.Copy(conn, pR)
		conn.Close()
	}()

	// No longer need these on this side.
	cR.Close()
	cW.Close()

	// TODO handle errors.
	cmd.Wait()

}
