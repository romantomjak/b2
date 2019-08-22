package command

import (
	"fmt"
	"strings"

	"github.com/romantomjak/b2/b2"
)

func (c *ListCommand) listFiles(path string) int {
	pathParts := strings.Split(path, "/")
	bucketName := pathParts[0]

	cmd := &b2.BucketListRequest{
		AccountID: c.Client.AccountID,
		Name:      bucketName,
	}
	buckets, _, err := c.Client.Bucket.List(cmd)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}
	if len(buckets) == 0 {
		c.Ui.Error(fmt.Sprintf("Bucket with name %q was not found.", bucketName))
		return 1
	}

	bucket := buckets[0]
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
