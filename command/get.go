package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
)

type GetCommand struct {
	Ui     cli.Ui
	Client *b2.Client
}

func (c *GetCommand) Help() string {
	helpText := `
Usage: b2 get <source> <destination>

  Downloads the given file to the destination.
`
	return strings.TrimSpace(helpText)
}

func (c *GetCommand) Synopsis() string {
	return "Download files"
}

func (c *GetCommand) Name() string { return "get" }

func (c *GetCommand) Run(args []string) int {
	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got both arguments
	args = flags.Args()
	numArgs := len(args)
	if numArgs != 2 {
		c.Ui.Error("This command takes two arguments: <source> and <destination>")
		return 1
	}

	// Create the destination file
	out, err := os.Create(args[1])
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	defer out.Close()

	// Write the data to file
	uri := fmt.Sprintf("%s/file/%s", c.Client.DownloadURL, args[0])
	_, err = c.Client.File.Download(uri, out)
	if err != nil {
		c.Ui.Error(err.Error())
		os.Remove(out.Name())
		return 1
	}

	c.Ui.Output(fmt.Sprintf("Downloaded %s to %s", args[0], args[1]))

	return 0
}
