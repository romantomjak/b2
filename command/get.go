package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/romantomjak/b2/b2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/sync/semaphore"
)

type GetCommand struct {
	*baseCommand
}

func (c *GetCommand) Help() string {
	helpText := `
Usage: b2 get <source> <destination>

  Downloads the given file to the destination.

General Options:

  ` + c.generalOptions()
	return strings.TrimSpace(helpText)
}

func (c *GetCommand) Synopsis() string {
	return "Download files"
}

func (c *GetCommand) Name() string { return "get" }

func (c *GetCommand) Run(args []string) int {
	flags := c.flagSet()
	flags.Usage = func() { c.ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got both arguments
	args = flags.Args()
	numArgs := len(args)
	if numArgs != 2 {
		c.ui.Error("This command takes two arguments: <source> and <destination>")
		return 1
	}

	// Resolve sources
	bucketName, filePrefix := splitBucketAndPrefix(args[0])

	ctx := context.TODO()

	// Create a client
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	bucket, err := findBucketByName(ctx, client, bucketName)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	req := &b2.FileListRequest{
		BucketID:  bucket.ID,
		Prefix:    filePrefix,
		Delimiter: "/",
	}

	files, _, err := client.File.List(ctx, req)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Resolve destination
	destination := args[1]
	if destination == "." {
		dir, err := os.Getwd()
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error: %v", err))
			return 1
		}
		destination = dir
	}

	// TODO: resolve ~/ paths

	return c.copy(bucketName, files, destination)
}

func (c *GetCommand) copy(bucketName string, sources []b2.File, destination string) int {
	if len(sources) == 0 {
		c.ui.Error("Error: source is not a file or directory")
		return 1
	}

	if len(sources) == 1 {
		_, err := os.Stat(destination)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			c.ui.Error(fmt.Sprintf("Error: %v", err))
			return 1
		}
	}

	// Multiple sources require the destination to be a folder.
	if len(sources) > 1 {
		info, err := os.Stat(destination)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				c.ui.Error(fmt.Sprintf("Error: %v", err))
				return 1
			}
			c.ui.Error("Error: destination directory does not exist")
			return 1
		}
		if !info.IsDir() {
			c.ui.Error(fmt.Sprintf("Error: %s is not a directory", destination))
			return 1
		}
	}

	maxWorkers := len(sources)
	if maxWorkers > 4 {
		maxWorkers = 4
	}

	p := mpb.New()

	sem := semaphore.NewWeighted(int64(maxWorkers))

	ctx := context.TODO()

	// Create a client
	client, err := c.Client()
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error: %v", err))
		return 1
	}

	for _, source := range sources {
		// TODO: ignore folders when copying without -R, write an error to stderr

		// Blocks until a worker becomes available
		if err := sem.Acquire(ctx, 1); err != nil {
			c.ui.Error(fmt.Sprintf("Error: failed to acquire semaphore: %v", err))
			break
		}

		go func(source b2.File) {
			defer sem.Release(1)

			bar := p.AddBar(int64(source.ContentLength),
				mpb.PrependDecorators(
					decor.Name(source.FileName),
				),
				mpb.AppendDecorators(
					decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
				),
			)

			filename := path.Join(destination, path.Base(source.FileName))

			// Create the destination file
			out, err := os.Create(filename)
			if err != nil {
				c.ui.Error(err.Error())
				return
			}
			defer out.Close()

			proxyWriter := bar.ProxyWriter(out)
			defer proxyWriter.Close()

			uri := fmt.Sprintf("%s/file/%s", client.DownloadURL, path.Join(bucketName, source.FileName))

			// See https://github.com/golang/go/issues/16474
			_, err = client.File.Download(ctx, uri, struct{ io.Writer }{proxyWriter})
			if err != nil {
				c.ui.Error(err.Error())
				return
			}
		}(source)
	}

	// Acquire all of the tokens to wait for any remaining workers to finish.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		c.ui.Error(fmt.Sprintf("Error: failed to acquire semaphore: %v", err))
		return 1
	}

	// Wait to flush the output
	p.Wait()

	return 0
}
