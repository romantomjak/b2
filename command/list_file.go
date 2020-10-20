package command

import (
	"fmt"
	"strings"

	"github.com/romantomjak/b2/b2"
)

func (c *ListCommand) listFiles(path string) int {
	pathParts := strings.SplitN(path, "/", 2)
	bucketName := pathParts[0]
	filePrefix := ""

	if len(pathParts) > 1 {
		filePrefix = pathParts[1]
	}

	bucket, err := c.findBucketByName(bucketName)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	req := &b2.FileListRequest{
		BucketID:  bucket.ID,
		Prefix:    filePrefix,
		Delimiter: "/",
	}

	files, _, err := client.File.List(req)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	for _, file := range files {
		c.ui.Output(file.FileName)
	}

	return 0
}

func (c *ListCommand) findBucketByName(name string) (*b2.Bucket, error) {
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	req := &b2.BucketListRequest{
		AccountID: client.AccountID,
		Name:      name,
	}

	buckets, _, err := client.Bucket.List(req)
	if err != nil {
		return nil, err
	}

	if len(buckets) == 0 {
		return nil, fmt.Errorf("bucket with name %q was not found", name)
	}

	return &buckets[0], nil
}
