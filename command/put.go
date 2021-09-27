package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/romantomjak/b2/b2"
)

type PutCommand struct {
	*baseCommand
}

func (c *PutCommand) Help() string {
	helpText := `
Usage: b2 put <source> <destination>

  Uploads the contents of source to destination. If destination
  contains a trailing slash it is treated as a directory and
  file is uploaded keeping the original filename.

General Options:

  ` + c.generalOptions()
	return strings.TrimSpace(helpText)
}

func (c *PutCommand) Synopsis() string {
	return "Upload files"
}

func (c *PutCommand) Name() string { return "put" }

func (c *PutCommand) Run(args []string) int {
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

	// Check that source file exists
	if !fileExists(args[0]) {
		c.ui.Error(fmt.Sprintf("File does not exist: %s", args[0]))
		return 1
	}

	// FIXME: remove when large file upload is implemented
	err := checkMaxFileSize(args[0])
	if err != nil {
		c.ui.Error("Large file upload is not yet implemented. Maximum file size is 100 MB")
		return 1
	}

	bucketName, filePrefix := destinationBucketAndFilename(args[0], args[1])

	// TODO: caching bucket name:id mappings could save this request
	bucket, err := c.findBucketByName(bucketName)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Create a client
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Request upload url
	ctx := context.TODO()

	uploadAuthReq := &b2.UploadAuthorizationRequest{
		BucketID: bucket.ID,
	}
	uploadAuth, _, err := client.File.UploadAuthorization(ctx, uploadAuthReq)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	_, _, err = client.File.Upload(ctx, uploadAuth, args[0], filePrefix)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	c.ui.Output(fmt.Sprintf("Uploaded %q to %q", args[0], path.Join(bucket.Name, filePrefix)))

	return 0
}

func (c *PutCommand) findBucketByName(name string) (*b2.Bucket, error) {
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	req := &b2.BucketListRequest{
		AccountID: client.Session.AccountID,
		Name:      name,
	}

	ctx := context.TODO()

	buckets, _, err := client.Bucket.List(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(buckets) == 0 {
		return nil, fmt.Errorf("bucket with name %q was not found", name)
	}

	return &buckets[0], nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// checkMaxFileSize checks that file is a "small" file
func checkMaxFileSize(filename string) error {
	var maxFileSize int64 = 100 << (10 * 2) // 100 mb

	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if info.Size() > maxFileSize {
		return errors.New("file is too big")
	}

	return nil
}

// destinationBucketAndFilename returns upload bucket and filePrefix
//
// b2 does not have a concept of folders, so if destination contains
// a trailing slash it is treated as a directory and file is uploaded
// keeping the original filename. If destination is simply a bucket
// name, it is asumed the destination is "/" and filename is preserved
func destinationBucketAndFilename(source, destination string) (string, string) {
	originalFilename := path.Base(source)

	destinationParts := strings.SplitN(destination, "/", 2)
	bucketName := destinationParts[0]
	filePrefix := ""

	if len(destinationParts) > 1 {
		if strings.HasSuffix(destinationParts[1], "/") {
			filePrefix = path.Join(destinationParts[1], originalFilename)
		} else {
			filePrefix = destinationParts[1]
		}
	}

	if filePrefix == "" {
		filePrefix = originalFilename
	}

	return bucketName, filePrefix
}
