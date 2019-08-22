package command

import (
	"fmt"
	"strings"

	"github.com/romantomjak/b2/b2"
)

func (c *ListCommand) listFiles(path string) int {
	pathParts := strings.Split(path, "/")
	bucketName := pathParts[0]

	bucket, err := c.findBucketByName(bucketName)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	cmd2 := &b2.FileListRequest{
		BucketID: bucket.ID,
	}
	files, _, err := c.Client.File.List(cmd2)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}
	for _, file := range files {
		c.Ui.Output(fmt.Sprintf("%s/%s", bucket.Name, file.FileName))
	}

	return 0
}

func (c *ListCommand) findBucketByName(name string) (*b2.Bucket, error) {
	req := &b2.BucketListRequest{
		AccountID: c.Client.AccountID,
		Name:      name,
	}

	buckets, _, err := c.Client.Bucket.List(req)
	if err != nil {
		return nil, err
	}

	if len(buckets) == 0 {
		return nil, fmt.Errorf("bucket with name %q was not found", name)
	}

	return &buckets[0], nil
}
