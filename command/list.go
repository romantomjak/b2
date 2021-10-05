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

  ` + c.generalOptions() + `

List Options:
  
  -long
    List files in long format. The following extra information
    is displayed for each file: file mode, number of bytes in
    the file and the timestamp of when file was uploaded.
`
	return strings.TrimSpace(helpText)
}

func (c *ListCommand) Synopsis() string {
	return "List files and buckets"
}

func (c *ListCommand) Name() string { return "list" }

func (c *ListCommand) Run(args []string) int {
	var longMode bool

	flags := c.flagSet()
	flags.Usage = func() { c.ui.Output(c.Help()) }
	flags.BoolVar(&longMode, "long", false, "List files in long mode")

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
		return c.listBuckets(longMode)
	}

	// User specified a path, so list files in path
	return c.listFiles(longMode, args[0])
}
