package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// HTTPClient interface can be satisfied by any http.Client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client manages communication with Backblaze API
type Client struct {
	// HTTP client used to communicate with the B2 API
	client HTTPClient

	// User agent for client
	UserAgent string

	// Base URL for API requests
	BaseURL *url.URL

	// Authorization token used for API calls
	Token string

	// The identifier for the account
	AccountID string

	// Services used for communicating with the API
	Bucket *BucketService
}

// NewClient returns a new Backblaze API client
func NewClient(httpClient HTTPClient) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:    httpClient,
		UserAgent: "b2/" + version.Version + " (+https://github.com/romantomjak/b2)",
		BaseURL:   baseURL,
	}

	c.Bucket = &BucketService{client: c}

	return c
}

// NewRequest creates an API request suitable for use with Client.Do
//
// The path should always be specified without a preceding slash. It will be
// resolved to the BaseURL of the Client.
//
// If specified, the value pointed to by body is JSON encoded and included in
// as the request body.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	if c.Token == "" {
		account, err := c.authorizeAccount()
		if err != nil {
			return nil, fmt.Errorf("authorization: %v", err)
		}

		err = c.reconfigureClient(account)
		if err != nil {
			return nil, err
		}
	}

	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.Token)

	return req, nil
}

// newRequest prepares a new Request
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

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

	req.Header.Add("User-Agent", c.UserAgent)

	return req, nil
}

// Do sends an API request and returns the API response
//
// The API response is JSON decoded and stored in the value pointed to by v
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = c.checkResponse(resp)
	if err != nil {
		return resp, fmt.Errorf("api: %v", err)
	}

	if v != nil {
		err := json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return nil, err
		}
	}

	return resp, err
}

// checkResponse checks the API response for errors and returns them if present
//
// Any code other than 2xx is an error, and the response will contain a JSON
// error structure indicating what went wrong
func (c *Client) checkResponse(r *http.Response) error {
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
	return fmt.Errorf("%v %v: %v %v", r.Request.Method, r.Request.URL, errResp.Code, errResp.Message)
}
