package command

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/romantomjak/b2/b2"
)

func (c *PutCommand) putLargeFile(info fs.FileInfo, src, dst string) int {
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

	// start large file upload
	// split file into n chunks using recommended part size
	// start uploading 4 parts
	// enqueue all other parts
	// wait until all chunks are uploaded
	// finish large file upload

	sha1, err := calculateFileSHA1(src)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	c.ui.Info(fmt.Sprintf("File sha1: %s", sha1))

	startLargeFileResp, err := c.startLargeFileUpload(ctx, sha1, info.ModTime(), bucket.ID, filePrefix)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	c.ui.Info(fmt.Sprintf("file id: %v", startLargeFileResp.FileID))

	partSHA1, err := c.uploadFileInChunks(ctx, info, src, startLargeFileResp.FileID)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		if _, err := c.cancelLargeFileUpload(ctx, startLargeFileResp.FileID); err != nil {
			c.ui.Error(fmt.Sprintf("Error: %v", err))
		}
		return 1
	}

	_, err = c.finishLargeFileUpload(ctx, startLargeFileResp.FileID, partSHA1)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	return 0
}

func (c *PutCommand) startLargeFileUpload(ctx context.Context, sha1 string, lastModified time.Time, bucketID, filename string) (*b2.File, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	req := &b2.StartLargeFileRequest{
		BucketID:    bucketID,
		Filename:    filename,
		ContentType: "b2/x-auto",
		FileInfo: map[string]string{
			"src_last_modified_millis": fmt.Sprintf("%d", lastModified.Unix()*1000),
			"large_file_sha1":          sha1,
		},
	}

	return client.File.StartLargeFile(ctx, req)
}

func (c *PutCommand) finishLargeFileUpload(ctx context.Context, fileID string, partSHA1 []string) (*b2.File, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	req := &b2.FinishLargeFileRequest{
		FileID:   fileID,
		PartSHA1: partSHA1,
	}

	return client.File.FinishLargeFile(ctx, req)
}

func (c *PutCommand) cancelLargeFileUpload(ctx context.Context, fileID string) (*b2.File, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	req := &b2.CancelLargeFileRequest{
		FileID: fileID,
	}

	return client.File.CancelLargeFile(ctx, req)
}

func (c *PutCommand) uploadFileInChunks(ctx context.Context, info fs.FileInfo, filename, fileID string) ([]string, error) {
	// Create a client
	// client, err := c.Client()
	// if err != nil {
	// 	return nil, err
	// }

	// partSize := client.RecommendedPartSize
	// FIXME: remove once testing is done
	partSize := int64(5000000)

	numParts := info.Size() / partSize
	maxParts := int64(10000)

	// Backblaze enforces a maximum limit of 10_000 parts
	if numParts > maxParts {
		partSize = info.Size() / maxParts
		numParts = maxParts
	}

	c.ui.Info(fmt.Sprintf("selected %v partsize, num chunks: %d", partSize, numParts))

	// Start workers
	numWorkers := 4
	chunks := make(chan chunk, numWorkers)
	results := make(chan *b2.FilePart)
	errors := make(chan error, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(chunks <-chan chunk, results chan<- *b2.FilePart) {
			for ch := range chunks {
				resp, err := c.uploadChunk(ctx, ch)
				if err != nil {
					errors <- err
					continue
				}
				results <- resp
			}
		}(chunks, results)
	}

	// Open file for reading.
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	offset := int64(0)

	// Backblaze wants part numbers to be contiguous numbers, starting with 1
	for i := int64(1); i <= numParts; i++ {
		i := i

		c.ui.Info(fmt.Sprintf("calculating hash for chunk %d...", i))

		hash := sha1.New()
		n, err := io.CopyN(hash, f, partSize)
		if err != nil {
			return nil, fmt.Errorf("compute part sha1 hash: %v", err)
		}
		partSHA1 := fmt.Sprintf("%x", hash.Sum(nil))

		c.ui.Info(fmt.Sprintf("chunk %d hash: %s", i, partSHA1))

		chunks <- chunk{filename, fileID, offset, i, n, partSHA1}

		c.ui.Info(fmt.Sprintf("enqueued chunk %d", i))

		offset += n
	}
	// note: not closing the chunks channel to allow for re-uploading

	partSHA1ByPartNumber := make(map[int]string)

outer:
	for {
		select {
		case err := <-errors:
			// FIXME: some errors can be retried
			return nil, err
		case p := <-results:
			partSHA1ByPartNumber[int(p.Number)] = p.ContentSHA1
			if len(partSHA1ByPartNumber) == int(numParts) {
				break outer
			}
		}
	}

	partSHA1s := make([]string, 0, len(results))
	for i := 1; i <= int(numParts); i++ {
		partSHA1s = append(partSHA1s, partSHA1ByPartNumber[i])
	}

	return partSHA1s, nil
}

func (c *PutCommand) uploadChunk(ctx context.Context, ch chunk) (*b2.FilePart, error) {
	c.ui.Info(fmt.Sprintf("chunk: %d, offset: %d, len: %d, sha1: %s", ch.partNum, ch.fileOffset, ch.partSize, ch.partSha1))

	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	// Request upload url
	uploadAuthReq := &b2.UploadPartAuthorizationRequest{
		FileID: ch.fileID,
	}
	uploadAuth, err := client.File.UploadPartAuthorization(ctx, uploadAuthReq)
	if err != nil {
		return nil, err
	}

	c.ui.Info(fmt.Sprintf("chunk upload url: %s", uploadAuth.UploadURL))

	// Open file for reading.
	f, err := os.Open(ch.filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c.ui.Info(fmt.Sprintf("seeking to %d to read %d bytes", ch.fileOffset, ch.partSize))

	f.Seek(ch.fileOffset, 0)
	r := io.LimitReader(f, ch.partSize)

	// Upload the part
	uploadReq := &b2.UploadPartRequest{
		Authorization: uploadAuth,
		PartNumber:    ch.partNum,
		Body:          r,
		ChecksumSHA1:  ch.partSha1,
		ContentLength: ch.partSize,
	}

	part, err := client.File.UploadPart(ctx, uploadReq)
	if err != nil {
		return nil, err
	}

	c.ui.Info(fmt.Sprintf("chunk %d uploaded", ch.partNum))

	return part, nil
}

type chunk struct {
	filename   string
	fileID     string
	fileOffset int64
	partNum    int64
	partSize   int64
	partSha1   string
}

func calculateFileSHA1(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
