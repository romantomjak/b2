package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
)

type ListCommand struct {
	Ui     cli.Ui
	Client *b2.Client
}

func (c *ListCommand) Help() string {
	helpText := `
Usage: b2 list [<path>]

  Lists files and buckets associated with an account.
`
	return strings.TrimSpace(helpText)
}

func (c *ListCommand) Synopsis() string {
	return "List files and buckets"
}

func (c *ListCommand) Name() string { return "list" }

func (c *ListCommand) Run(args []string) int {
	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we either got none or exactly 1 argument
	args = flags.Args()
	numArgs := len(args)
	if numArgs > 1 {
		c.Ui.Error("This command takes one argument: <path>")
		return 1
	}

	// List buckets
	cmd := &b2.BucketListRequest{
		AccountID: c.Client.AccountID,
	}
	buckets, _, err := c.Client.Bucket.List(cmd)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}
	for _, bucket := range buckets {
		c.Ui.Output(bucket.Name)
	}

	return 0
}
