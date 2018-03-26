package config

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Configs struct {
	KubeConfig string
	Port       string
}

func initVariable() {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "2")
	flag.CommandLine.Parse([]string{})
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.String("kubeconfig", "", "absolute path of the kubeconfig file")
	pflag.String("port", "3000", "serve port")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

func NewConfig() *Configs {
	initVariable()
	return &Configs{
		Port:       ":" + viper.GetString("port"),
		KubeConfig: viper.GetString("kubeconfig"),
	}
}
