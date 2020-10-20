package command

import (
	"fmt"

	"github.com/romantomjak/b2/b2"
)

func (c *ListCommand) listBuckets() int {
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	req := &b2.BucketListRequest{
		AccountID: client.AccountID,
	}

	buckets, _, err := client.Bucket.List(req)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	for _, bucket := range buckets {
		c.ui.Output(bucket.Name + "/")
	}

	return 0
}
