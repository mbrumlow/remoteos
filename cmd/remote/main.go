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

	log.SetOutput(os.Stderr)

	lh := remote.NewLocalHost()

	out := os.Stdout
	os.Stdout = os.Stderr

	lh.Run(os.Stdin, out)
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
	cmd.Stdout = cW
	cmd.Stdin = cR
	cmd.Stderr = os.Stdout

	log.Printf("Got new connection, forking...\n")

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to fork: %v\n", err)
		return
	}

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
