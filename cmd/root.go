// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	goflag "flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/tony24681379/k8s-alert-controller/server"
)

type options struct {
	Kubeconfig string
	Port       string
}

var opts = options{}

func init() {
	goflag.Set("alsologtostderr", "true")
	goflag.CommandLine.Parse([]string{})
}

// RootCmd represents the base command when called without any subcommands
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s-tools",
		Short: "Kubernetes tools",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flags.StringVar(&opts.Port, "port", "10100", "serve port")
	cmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	return cmd
}

// Run init cobra command
func Run() error {
	rootCmd := newRootCmd()
	return rootCmd.Execute()
}

func runServer(opts options) {
	err := server.Server(opts.Kubeconfig, opts.Port)
	if err != nil {
		glog.Fatal(err)
		panic(err)
	}
}
