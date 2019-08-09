package client

import (
	"net/http"
	"net/url"

	"github.com/romantomjak/b2/version"
)

const (
	defaultBaseURL      = "https://api.backblazeb2.com/"
	authorizeAccountURL = "b2api/v2/b2_authorize_account"
)

// ApplicationCredentials are used to authorize the client
type ApplicationCredentials struct {
	// The ID of the key
	KeyID string

	// The secret part of the key
	KeySecret string
}

// Client manages communication with Backblaze API
type Client struct {
	// HTTP client used to communicate with the B2 API
	client *http.Client

	// Credentials for authorizing the client
	credentials *ApplicationCredentials

	// User agent for client
	UserAgent string

	// Base URL for API requests
	BaseURL *url.URL

	// Authorization token used for API calls
	Token string
}

// NewClient returns a new Backblaze API client
func NewClient(credentials *ApplicationCredentials) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		client:      http.DefaultClient,
		credentials: credentials,
		UserAgent:   "b2/" + version.Version + " (+https://github.com/romantomjak/b2)",
		BaseURL:     baseURL,
	}
}

// NewRequest creates an API request suitable for use with Client.Do
//
// The path should always be specified without a preceding slash. It will be
// resolved to the BaseURL of the Client.
func (c *Client) NewRequest(method, path string) (*http.Request, error) {
	req, err := c.newRequest(method, path)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.Token)

	return req, nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
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
