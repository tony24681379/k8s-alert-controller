package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/tony24681379/k8s-tools/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		glog.Fatal(err)
		os.Exit(-1)
	}
}
