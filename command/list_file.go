package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/romantomjak/b2/b2"
)

const (
	TB = 1000000000000
	GB = 1000000000
	MB = 1000000
	KB = 1000
)

func (c *ListCommand) listFiles(longMode bool, path string) int {
	pathParts := strings.SplitN(path, "/", 2)
	bucketName := pathParts[0]
	filePrefix := ""

	if len(pathParts) > 1 {
		filePrefix = pathParts[1]
	}

	bucket, err := c.findBucketByName(bucketName)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	req := &b2.FileListRequest{
		BucketID:  bucket.ID,
		Prefix:    filePrefix,
		Delimiter: "/",
	}

	ctx := context.TODO()

	files, _, err := client.File.List(ctx, req)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	for _, file := range files {
		if longMode {
			size := lenReadable(file.ContentLength, 1)

			t := time.UnixMilli(file.UploadTimestamp)
			ts := t.Format("_2 Jan 15:04")
			if time.Now().Sub(t).Hours()/24 > 180 {
				ts = t.Format("_2 Jan 2006")
			}

			mode := "-"
			if strings.HasSuffix(file.FileName, "/") {
				mode = "d"
			}

			c.ui.Output(fmt.Sprintf("%s  %6s %12s %s", mode, size, ts, file.FileName))
		} else {
			c.ui.Output(file.FileName)
		}
	}

	return 0
}

func (c *ListCommand) findBucketByName(name string) (*b2.Bucket, error) {
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	req := &b2.BucketListRequest{
		AccountID: client.AccountID,
		Name:      name,
	}

	ctx := context.TODO()

	buckets, _, err := client.Bucket.List(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(buckets) == 0 {
		return nil, fmt.Errorf("bucket with name %q was not found", name)
	}

	return &buckets[0], nil
}

func lenReadable(length int, decimals int) (out string) {
	var unit string
	var i int
	var remainder int

	// Get whole number, and the remainder for decimals
	if length > TB {
		unit = "T"
		i = length / TB
		remainder = length - (i * TB)
	} else if length > GB {
		unit = "G"
		i = length / GB
		remainder = length - (i * GB)
	} else if length > MB {
		unit = "M"
		i = length / MB
		remainder = length - (i * MB)
	} else if length > KB {
		unit = "K"
		i = length / KB
		remainder = length - (i * KB)
	} else {
		return strconv.Itoa(length) + "B"
	}

	if decimals == 0 {
		return strconv.Itoa(i) + unit
	}

	// This is to calculate missing leading zeroes
	width := 0
	if remainder > GB {
		width = 12
	} else if remainder > MB {
		width = 9
	} else if remainder > KB {
		width = 6
	} else {
		width = 3
	}

	// Insert missing leading zeroes
	remainderString := strconv.Itoa(remainder)
	for iter := len(remainderString); iter < width; iter++ {
		remainderString = "0" + remainderString
	}
	if decimals > len(remainderString) {
		decimals = len(remainderString)
	}

	return fmt.Sprintf("%d.%s%s", i, remainderString[:decimals], unit)
}
