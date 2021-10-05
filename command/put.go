package command

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"

	"github.com/romantomjak/b2/b2"
)

const (
	ClearLine    = "\033[2K"
	MoveCursorUp = "\033[1F"
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

	file, err := os.Open(args[0])
	if err != nil {
		c.ui.Error(fmt.Sprintf("Cannot open file: %s", args[0]))
		return 1
	}

	info, err := file.Stat()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Cannot open file: %s", args[0]))
		return 1
	}

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Cannot calculate file hash: %s", err))
		return 1
	}
	sha1 := fmt.Sprintf("%x", hash.Sum(nil))

	file.Seek(0, 0)

	progress := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	bar := progress.Add(info.Size(),
		mpb.NewBarFiller(mpb.BarStyle().Rbound("|")),
		mpb.PrependDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)

	proxyReader := bar.ProxyReader(file)
	defer proxyReader.Close()

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

	_, _, err = client.File.Upload(ctx, &b2.UploadInput{
		Authorization: uploadAuth,
		Body:          proxyReader,
		Key:           filePrefix,
		ContentSHA1:   sha1,
		ContentLength: info.Size(),
		Metadata: map[string]string{
			"src_last_modified_millis": fmt.Sprintf("%d", info.ModTime().Unix()*1000),
		},
	})
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	progress.Wait()

	// Delete the progress bar line
	c.ui.Output(strings.Join([]string{ClearLine, MoveCursorUp, ClearLine, MoveCursorUp}, ";"))

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
