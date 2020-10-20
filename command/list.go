package command

import (
	"strings"
)

type ListCommand struct {
	*baseCommand
}

func (c *ListCommand) Help() string {
	helpText := `
Usage: b2 list [<path>]

  Lists files and buckets associated with an account.

General Options:

  ` + c.generalOptions()
	return strings.TrimSpace(helpText)
}

func (c *ListCommand) Synopsis() string {
	return "List files and buckets"
}

func (c *ListCommand) Name() string { return "list" }

func (c *ListCommand) Run(args []string) int {
	flags := c.flagSet()
	flags.Usage = func() { c.ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we either got none or exactly 1 argument
	args = flags.Args()
	numArgs := len(args)
	if numArgs > 1 {
		c.ui.Error("This command takes one argument: <path>")
		return 1
	}

	// No path argument - list buckets
	if numArgs == 0 {
		return c.listBuckets()
	}

	// User specified a path, so list files in path
	return c.listFiles(args[0])
}
