package main

import (
	color "github.com/TwiN/go-color"
	"github.com/mvazquezc/cluster0-operators/cmd/cli"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func main() {
	command := newRootCommand()
	if err := command.Execute(); err != nil {
		log.Fatalf(color.InRed("[ERROR] ")+"%s", err.Error())
	}

}

// newRootCommand implements the root command of example-ci
func newRootCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "cluster0-operators",
		Short: "Get your cluster0 configured",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	c.AddCommand(cli.NewRunCommand())
	c.AddCommand(cli.NewVersionCommand())

	return c
}
