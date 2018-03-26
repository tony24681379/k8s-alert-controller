package main

import (
	"github.com/golang/glog"
	"github.com/tony24681379/k8s-alert-controller/config"
	"github.com/tony24681379/k8s-alert-controller/server"
)

func main() {
	configs := config.NewConfig()
	err := server.Server(configs.KubeConfig, configs.Port)
	if err != nil {
		glog.Fatal(err)
		panic(err)
	}
}
