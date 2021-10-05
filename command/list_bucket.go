package command

import (
	"context"
	"fmt"

	"github.com/romantomjak/b2/b2"
)

func (c *ListCommand) listBuckets(longMode bool) int {
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	req := &b2.BucketListRequest{
		AccountID: client.AccountID,
	}

	ctx := context.TODO()

	buckets, _, err := client.Bucket.List(ctx, req)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	for _, bucket := range buckets {
		if longMode {
			c.ui.Output(fmt.Sprintf("d  0B %s", bucket.Name+"/"))
		} else {
			c.ui.Output(bucket.Name + "/")
		}
	}

	return 0
}
