package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/romantomjak/b2/version"
)

const (
	defaultBaseURL      = "https://api.backblazeb2.com/"
	authorizeAccountURL = "b2api/v2/b2_authorize_account"
)

// authorizeAccount represents the authorization response from the B2 API
type authorizeAccount struct {
	AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	AccountID               string `json:"accountId"`
	Allowed                 struct {
		BucketID     string      `json:"bucketId"`
		BucketName   string      `json:"bucketName"`
		Capabilities []string    `json:"capabilities"`
		NamePrefix   interface{} `json:"namePrefix"`
	} `json:"allowed"`
	APIURL              string `json:"apiUrl"`
	AuthorizationToken  string `json:"authorizationToken"`
	DownloadURL         string `json:"downloadUrl"`
	RecommendedPartSize int    `json:"recommendedPartSize"`
}

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

	// The identifier for the account
	AccountID string
}

// NewClient returns a new Backblaze API client
func NewClient(credentials *ApplicationCredentials) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	client := &Client{
		client:      http.DefaultClient,
		credentials: credentials,
		UserAgent:   "b2/" + version.Version + " (+https://github.com/romantomjak/b2)",
		BaseURL:     baseURL,
	}

	return client
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
		account, tokenErr := c.authorizeAccount()
		if tokenErr != nil {
			return nil, tokenErr
		}

		accountErr := c.reconfigureClient(account)
		if accountErr != nil {
			return nil, accountErr
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
		err = json.NewEncoder(buf).Encode(body)
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

// authorizeAccount is used to log in to the B2 API
//
// This must be the very first API call to obtain essential account information
func (c *Client) authorizeAccount() (*authorizeAccount, error) {
	req, err := c.newRequest(http.MethodGet, authorizeAccountURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.credentials.KeyID, c.credentials.KeySecret)

	account := new(authorizeAccount)
	_, sendErr := c.Do(req, &account)
	if sendErr != nil {
		return nil, sendErr
	}

	return account, nil
}

// Do sends an API request and returns the API response.
//
// The API response is JSON decoded and stored in the value pointed to by v.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return nil, err
		}
	}
	return resp, err
}

// reconfigureClient is used to configure the client after authentication
//
// Authorization API call returns a token and a URL that should be used as
// the base URL for subsequent API calls
func (c *Client) reconfigureClient(account *authorizeAccount) error {
	c.Token = account.AuthorizationToken
	c.AccountID = account.AccountID

	newBaseURL, err := url.Parse(account.APIURL)
	if err != nil {
		return err
	}
	c.BaseURL = newBaseURL

	return nil
}
