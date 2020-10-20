package command

import (
	"flag"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
)

// baseCommand is embedded in all commands to provide common logic
type baseCommand struct {
	ui     cli.Ui
	client *b2.Client

	// Whether to disable disk cache
	noCache bool

	keyId     string
	keySecret string
}

func (c *baseCommand) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("Global Options", flag.ContinueOnError)
	fs.StringVar(&c.keyId, "key-id", "", "")
	fs.StringVar(&c.keySecret, "key-secret", "", "")
	fs.BoolVar(&c.noCache, "no-cache", false, "")

	// try to get credentials from environment
	if c.keyId == "" {
		c.keyId = os.Getenv("B2_KEY_ID")
	}
	if c.keySecret == "" {
		c.keySecret = os.Getenv("B2_KEY_SECRET")
	}

	// check if cache is disabled via environment
	if c.noCache == false {
		_, found := os.LookupEnv("B2_NO_CACHE")
		if found {
			c.noCache = true
		}
	}

	return fs
}

func (c *baseCommand) generalOptions() string {
	helpText := `
  -key-id
    The ID of the application key. Overrides the
    B2_KEY_ID environment variable if set.

  -key-secret
    The secret part of the application key. Overrides
    the B2_KEY_SECRET environment variable if set.

  -no-cache
    Disable disk cache. This is highly not recommended, but
    might come handy when leaving a trace on disk is unwise.
    Alternatively, B2_NO_CACHE may be set.
`
	return strings.TrimSpace(helpText)
}

func (c *baseCommand) Client() (*b2.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	opts := []b2.ClientOpt{}

	// disk cache is disabled, so we'll replace it with an
	// in-memory based cache
	if c.noCache {
		cache, err := b2.NewInMemoryCache()
		if err != nil {
			return nil, err
		}
		opts = append(opts, b2.SetCache(cache))
	}

	client, err := b2.NewClient(c.keyId, c.keySecret, opts...)
	if err != nil {
		return nil, err
	}

	c.client = client

	return c.client, nil
}
