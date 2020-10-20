package command

import (
	"fmt"
	"os"
	"strings"
)

type GetCommand struct {
	*baseCommand
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
	flags := c.flagSet()
	flags.Usage = func() { c.ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got both arguments
	args = flags.Args()
	numArgs := len(args)
	if numArgs != 2 {
		c.ui.Error("This command takes two arguments: <source> and <destination>")
		return 1
	}

	// Create the destination file
	out, err := os.Create(args[1])
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	defer out.Close()

	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Write the data to file
	uri := fmt.Sprintf("%s/file/%s", client.DownloadURL, args[0])
	_, err = client.File.Download(uri, out)
	if err != nil {
		c.ui.Error(err.Error())
		os.Remove(out.Name())
		return 1
	}

	c.ui.Output(fmt.Sprintf("Downloaded %s to %s", args[0], args[1]))

	return 0
}
