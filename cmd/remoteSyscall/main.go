package main

import (
	remote "github.com/mbrumlow/remote/pkg"
)

func main() {
	lh := remote.NewLocalHost(":7575")
	lh.Run()
}
