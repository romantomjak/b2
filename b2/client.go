package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/romantomjak/b2/version"
)

const (
	defaultBaseURL = "https://api.backblazeb2.com/"
)

// An errorResponse contains the error caused by an API request
type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Client manages communication with Backblaze API
type Client struct {
	// HTTP client used to communicate with the B2 API
	client *http.Client

	// User agent for client
	userAgent string

	// Base URL for API requests
	baseURL *url.URL

	// API authorization data
	auth *authorization

	// The identifier for the account
	AccountID string

	// The base URL for downloading files
	DownloadURL *url.URL

	// Services used for communicating with the API
	Bucket *BucketService
	File   *FileService
}

// ClientOpt are options for New
type ClientOpt func(*Client) error

// NewClient returns a new Backblaze API client
func NewClient(opts ...ClientOpt) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:    http.DefaultClient,
		baseURL:   baseURL,
		userAgent: "b2/" + version.Version + " (+https://github.com/romantomjak/b2)",
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	err := c.authorize()
	if err != nil {
		return nil, fmt.Errorf("authorization: %v", err)
	}

	c.AccountID = c.auth.AccountID

	c.Bucket = &BucketService{client: c}
	c.File = &FileService{client: c}

	return c, nil
}

// SetBaseURL is a client option for setting the base URL
func SetBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}
		c.baseURL = u
		return nil
	}
}

// NewRequest creates an API request suitable for use with Client.Do
//
// The path should always be specified without a preceding slash. It will be
// resolved to the BaseURL of the Client.
//
// If specified, the value pointed to by body is JSON encoded and included in
// as the request body.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.auth.AuthorizationToken)

	return req, nil
}

// newRequest prepares a new Request
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.userAgent)

	return req, nil
}

// Do sends an API request and returns the API response
//
// The API response is JSON decoded and stored in the value pointed to by v.
// If v implements the io.Writer interface, the raw response will be written
// to v, without attempting to decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return nil, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err := json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, err
}

// checkResponse checks the API response for errors and returns them if present
//
// Any code other than 2xx is an error, and the response will contain a JSON
// error structure indicating what went wrong
func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("%v %v: empty error body", r.Request.Method, r.Request.URL)
	}

	errResp := new(errorResponse)
	err = json.Unmarshal(data, errResp)
	if err != nil {
		errResp.Message = string(data)
	}

	if r.StatusCode == 401 && errResp.Code == "expired_auth_token" {
		return ErrExpiredToken
	}

	return fmt.Errorf("%v %v: %v %v", r.Request.Method, r.Request.URL, errResp.Code, errResp.Message)
}
