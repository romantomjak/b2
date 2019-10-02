package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/command"
	"github.com/romantomjak/b2/version"
)

func main() {
	os.Exit(Run(os.Stdin, os.Stdout, os.Stdout, os.Args[1:]))
}

func Run(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	ui := &cli.BasicUi{
		Reader:      stdin,
		Writer:      stdout,
		ErrorWriter: stderr,
	}

	client, err := b2.NewClient()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err.Error())
		return 1
	}

	c := cli.NewCLI("b2", version.Version)
	c.Args = args
	c.Commands = map[string]cli.CommandFactory{
		"create": func() (cli.Command, error) {
			return &command.CreateBucketCommand{
				Ui:     ui,
				Client: client,
			}, nil
		},
		"list": func() (cli.Command, error) {
			return &command.ListCommand{
				Ui:     ui,
				Client: client,
			}, nil
		},
		"get": func() (cli.Command, error) {
			return &command.GetCommand{
				Ui:     ui,
				Client: client,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Ui:      ui,
				Version: version.FullVersion(),
			}, nil
		},
	}

	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}
