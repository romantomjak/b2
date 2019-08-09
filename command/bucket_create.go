package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	b2 "github.com/romantomjak/b2/client"
)

type CreateBucketCommand struct {
	Ui     cli.Ui
	Client *b2.Client
}

func (c *CreateBucketCommand) Help() string {
	helpText := `
Usage: b2 create [options] <bucket-name>

  Creates a new bucket belonging to the account used to create it. The name
  must be globally unique and there is a limit of 100 buckets per account.

  Options:

    -type
      Either "public", meaning that files in this bucket can be downloaded by
      anybody, or "private", meaning that you need an authorization token to
      download the files.
`
	return strings.TrimSpace(helpText)
}

func (c *CreateBucketCommand) Synopsis() string {
	return "Create a new bucket"
}

func (c *CreateBucketCommand) Name() string { return "create" }

func (c *CreateBucketCommand) Run(args []string) int {
	var bucketType string

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&bucketType, "type", "private", "Change bucket type")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got only one argument
	args = flags.Args()
	if l := len(args); l != 1 {
		c.Ui.Error("This command takes one argument: <bucket-name>")
		return 1
	}

	// Validate bucket type
	if bucketType != "public" && bucketType != "private" {
		c.Ui.Error(`-type must be either "public" or "private"`)
		return 1
	}

	// Get the bucket name
	bucketName := args[0]

	c.Ui.Output(fmt.Sprintf("Successfully created %q Bucket!", bucketName))

	return 0
}
