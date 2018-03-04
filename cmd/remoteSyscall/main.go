package main

import "github.com/mbrumlow/remoteos/pkg/remote"

func main() {
	lh := remote.NewLocalHost(":7575")
	lh.Run()
}
