package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"

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
	info, err := os.Stat(args[0])
	if err != nil {
		if os.IsNotExist(err) {
			c.ui.Error(fmt.Sprintf("File does not exist: %s", args[0]))
			return 1
		}
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Create a client
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	if info.Size() > client.RecommendedPartSize {
		return c.putLargeFile(info, args[0], args[1])
	}

	return c.putSmallFile(info, args[0], args[1])
}

// progressReader is a helper for tracking the amount of bytes uploaded.
type progressReader struct {
	file     *os.File
	progress *uiprogress.Progress
	bar      *uiprogress.Bar
}

func newProgressReader(f *os.File) (*progressReader, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	progress := uiprogress.New()
	progress.SetRefreshInterval(time.Millisecond * 1)

	bar := progress.AddBar(int(info.Size()))
	bar.AppendCompleted()
	bar.PrependElapsed()

	return &progressReader{
		file:     f,
		progress: progress,
		bar:      bar,
	}, nil
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.file.Read(p)
	for i := 0; i < n; i++ {
		pr.bar.Incr()
	}
	return
}

func (pr *progressReader) Start() {
	pr.progress.Start()
}

func (pr *progressReader) Stop() {
	pr.progress.Stop()
}

func (c *PutCommand) findBucketByName(name string) (*b2.Bucket, error) {
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	req := &b2.BucketListRequest{
		AccountID: client.AccountID,
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
