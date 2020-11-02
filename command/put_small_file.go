package command

import (
	"context"
	"fmt"
	"path"

	"github.com/romantomjak/b2/b2"
)

func (c *PutCommand) putSmallFile(source, destination string) int {
	// Check that source file exists
	if !fileExists(source) {
		c.ui.Error(fmt.Sprintf("File does not exist: %s", source))
		return 1
	}

	// FIXME: remove when large file upload is implemented
	err := checkMaxFileSize(source)
	if err != nil {
		c.ui.Error("Large file upload is not yet implemented. Maximum file size is 100 MB")
		return 1
	}

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
