package main

import (
	"log"
	"os"

	"github.com/mbrumlow/remoteos/pkg/remote"
)

func main() {

	log.Printf("PATH: %v\n", os.Args[0])

	lh := remote.NewLocalHost(":7575")
	lh.Run()
}
