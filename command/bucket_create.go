package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/romantomjak/b2/b2"
)

type CreateBucketCommand struct {
	*baseCommand
}

func (c *CreateBucketCommand) Help() string {
	helpText := `
Usage: b2 create [options] <bucket-name>

  Creates a new bucket belonging to the account used to create it. The name
  must be globally unique and there is a limit of 100 buckets per account.

General Options:

  ` + c.generalOptions() + `

Create Options:

  -type
    Either "public", meaning that files in this bucket can be downloaded by
    anybody, or "private", meaning that you need an authorization token to
    download the files. Defaults to "private".
`
	return strings.TrimSpace(helpText)
}

func (c *CreateBucketCommand) Synopsis() string {
	return "Create a new bucket"
}

func (c *CreateBucketCommand) Name() string { return "create" }

func (c *CreateBucketCommand) Run(args []string) int {
	var bucketType string

	flags := c.flagSet()
	flags.Usage = func() { c.ui.Output(c.Help()) }
	flags.StringVar(&bucketType, "type", "private", "Change bucket type")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got only one argument
	args = flags.Args()
	if l := len(args); l != 1 {
		c.ui.Error("This command takes one argument: <bucket-name>")
		return 1
	}

	// Validate bucket type
	if bucketType != "public" && bucketType != "private" {
		c.ui.Error(`-type must be either "public" or "private"`)
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Create the bucket
	b := &b2.BucketCreateRequest{
		AccountID: client.Session.AccountID,
		Name:      args[0],
		Type:      "all" + strings.Title(bucketType),
	}

	ctx := context.TODO()

	bucket, _, err := client.Bucket.Create(ctx, b)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	c.ui.Output(fmt.Sprintf("Bucket %q created with ID %q", bucket.Name, bucket.ID))

	return 0
}
