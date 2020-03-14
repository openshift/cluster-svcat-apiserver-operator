package main

import (
	goflag "flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/cmd/operator"
	"github.com/openshift/cluster-svcat-apiserver-operator/pkg/version"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
)

var versionFlag = goflag.Bool("version", false, "displays source commit info.")

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	goflag.Parse()
	if *versionFlag {
		fmt.Print(version.Commit())
		// Exit immediately
		os.Exit(0)
	}

	command := NewSSCSCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}

func NewSSCSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster-svcat-apiserver-operator",
		Short: "OpenShift cluster svcat-apiserver operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	cmd.AddCommand(operator.NewOperator())

	return cmd
}
