package command

import (
	"fmt"

	"github.com/romantomjak/b2/b2"
)

func (c *ListCommand) listBuckets() int {
	req := &b2.BucketListRequest{
		AccountID: c.Client.AccountID,
	}

	buckets, _, err := c.Client.Bucket.List(req)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	for _, bucket := range buckets {
		c.Ui.Output(bucket.Name + "/")
	}

	return 0
}
