package command

import (
	"context"
	"fmt"
	"path"

	"github.com/romantomjak/b2/b2"
)

func (c *PutCommand) putSmallFile(source, destination string) int {
	bucketName, filePrefix := destinationBucketAndFilename(source, destination)

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

	_, _, err = client.File.Upload(ctx, uploadAuth, source, filePrefix)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	c.ui.Output(fmt.Sprintf("Uploaded %q to %q", source, path.Join(bucket.Name, filePrefix)))

	return 0
}
