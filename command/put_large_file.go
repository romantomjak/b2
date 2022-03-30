package command

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/romantomjak/b2/b2"
)

func (c *PutCommand) putLargeFile(source, destination string) int {
	bucketName, filePrefix := destinationBucketAndFilename(source, destination)

	// TODO: caching bucket name:id mappings could save this request
	bucket, err := c.findBucketByName(bucketName)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	f, err := os.Open(source)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	hash := sha1.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	sha1sum := fmt.Sprintf("%x", hash.Sum(nil))
	f.Seek(0, 0)

	// Create a client
	client, err := c.Client()
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

	ctx := context.TODO()
	req := &b2.LargeFileRequest{
		BucketID:    bucket.ID,
		Filename:    filePrefix,
		ContentType: "b2/x-auto",
		FileInfo: map[string]string{
			"src_last_modified_millis": fmt.Sprintf("%d", info.ModTime().Unix()*1000),
			"large_file_sha1":          sha1sum,
		},
	}
	file, _, err := client.File.StartLargeFile(ctx, req)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	partSize := client.RecommendedPartSize
	maxParts := int64(10000)
	numParts := info.Size() / partSize

	// Backblaze enforces a maximum limit of 10_000 parts
	if numParts > maxParts {
		partSize = info.Size() / maxParts
		numParts = maxParts
		// TODO: return error instead?
	}

	c.ui.Info(fmt.Sprintf("Selected %d parts of %d", numParts, partSize))

	chunks := make(chan chunk)
	results := make(chan chunk)
	for i := 0; i < 4; i++ {
		go c.uploadChunk(chunks, results)
	}

	partSHA1s := make([]string, 0, numParts)

	// Backblaze wants part numbers to be contiguous numbers, starting with 1
	for i := int64(1); i < numParts; i++ {
		ch := chunk{source, file.FileID, 0, i, partSize, nil}
		partSHA1, err := chunkSHA1(ch)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error: %v", err))
			return 1
		}
		partSHA1s = append(partSHA1s, partSHA1)
		chunks <- ch
	}

	// TODO: range over results chan and re-enque failed chunks

	ctx = context.TODO()
	finreq := &b2.FinishLargeFileRequest{
		FileID:    file.FileID,
		PartSHA1s: partSHA1s,
	}
	_, _, err = client.File.FinishLargeFile(ctx, finreq)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	c.ui.Output(fmt.Sprintf("Uploaded %q to %q", source, path.Join(bucket.Name, filePrefix)))

	return 0
}

type chunk struct {
	filename   string
	fileID     string
	fileOffset int64
	partNum    int64
	partSize   int64
	err        error
}

func chunkSHA1(c chunk) (string, error) {
	f, err := os.Open(c.filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	f.Seek(c.fileOffset, 0)

	r := io.LimitReader(f, c.partSize)

	hash := sha1.New()
	_, err = io.Copy(hash, r)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (c *PutCommand) uploadChunk(chunks <-chan chunk, results chan<- chunk) {
	var uploadAuth *b2.UploadAuthorization

	for chunk := range chunks {
		sha1sum, err := chunkSHA1(chunk)
		if err != nil {
			chunk.err = err
			results <- chunk
			continue
		}

		f, err := os.Open(chunk.filename)
		if err != nil {
			chunk.err = err
			results <- chunk
			continue
		}

		f.Seek(chunk.fileOffset, 0)
		r := io.LimitReader(f, chunk.partSize)

		client, err := c.Client()
		if err != nil {
			chunk.err = err
			results <- chunk
			f.Close()
			continue
		}

		if uploadAuth == nil {
			ctx := context.TODO()
			req := &b2.PartUploadAuthorizationRequest{
				FileID: chunk.fileID,
			}
			uploadAuth, _, err = client.File.PartUploadAuthorization(ctx, req)
			if err != nil {
				chunk.err = err
				results <- chunk
				// TODO: retry?
				f.Close()
				continue
			}
		}

		ctx := context.TODO()
		req := &b2.UploadRequest{
			Body:          r,
			ContentSHA1:   sha1sum,
			Authorization: uploadAuth,
			PartNumber:    chunk.partNum,
		}
		_, _, err = client.File.UploadPart(ctx, req)
		if err != nil {
			chunk.err = err
			results <- chunk
			f.Close()
			continue
		}

		// Clear err (in case of retries)
		chunk.err = nil
		results <- chunk
	}
}
