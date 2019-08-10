package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/cli"
	b2 "github.com/romantomjak/b2/api"
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

	b2client := b2.NewClient(&b2.ApplicationCredentials{
		KeyID:     os.Getenv("B2_KEY_ID"),
		KeySecret: os.Getenv("B2_KEY_SECRET"),
	})

	c := cli.NewCLI("b2", version.Version)
	c.Args = args
	c.Commands = map[string]cli.CommandFactory{
		"create": func() (cli.Command, error) {
			return &command.CreateBucketCommand{
				Ui:     ui,
				Client: b2client,
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
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}
