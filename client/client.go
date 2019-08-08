package client

import (
	"net/http"
	"net/url"

	"github.com/romantomjak/b2/version"
)

// Client manages communication with Backblaze API
type Client struct {
	// HTTP client used to communicate with the B2 API
	client *http.Client

	// User agent for client
	UserAgent string

	// Base URL for API requests
	BaseURL *url.URL
}

// NewClient returns a new Backblaze API client
func NewClient() *Client {
	// This will be replaced with a new URL returned by the
	// account authorization API call.
	baseURL, _ := url.Parse("https://api.backblazeb2.com/")

	return &Client{
		client:    http.DefaultClient,
		UserAgent: "b2/" + version.Version + " (+https://github.com/romantomjak/b2)",
		BaseURL:   baseURL,
	}
}

// NewRequest creates an API request suitable for use with Client.Do
//
// The path should always be specified without a preceding slash. It will be
// resolved to the BaseURL of the Client.
//
// If specified, the value pointed to by body is JSON encoded and included
// in as the request body.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.UserAgent)

	return req, nil
}
