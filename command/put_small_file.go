package command

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/romantomjak/b2/b2"
)

func (c *PutCommand) putSmallFile(info fs.FileInfo, src, dst string) int {
	bucketName, filePrefix := destinationBucketAndFilename(src, dst)

	// Create a client
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	ctx := context.TODO()

	// TODO: caching bucket name:id mappings could save this request
	bucket, err := findBucketByName(ctx, client, bucketName)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Request upload url
	uploadAuthReq := &b2.UploadAuthorizationRequest{
		BucketID: bucket.ID,
	}
	uploadAuth, _, err := client.File.UploadAuthorization(ctx, uploadAuthReq)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Open file for reading.
	f, err := os.Open(src)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	defer f.Close()

	// Calculate SHA1 checksum
	hash := sha1.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	sha1 := fmt.Sprintf("%x", hash.Sum(nil))

	// Rewind the file
	f.Seek(0, 0)

	// Create a progress bar.
	pr, err := newProgressReader(f)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	pr.Start()
	defer pr.Stop()

	uploadReq := &b2.UploadRequest{
		Authorization: uploadAuth,
		Body:          pr,
		Key:           filePrefix,
		ChecksumSHA1:  sha1,
		ContentLength: info.Size(),
		LastModified:  info.ModTime(),
	}

	_, _, err = client.File.Upload(ctx, uploadReq)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	return 0
}
