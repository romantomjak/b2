package b2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/romantomjak/b2/version"
)

const (
	defaultBaseURL   = "https://api.backblazeb2.com/"
	authorizationURL = "b2api/v2/b2_authorize_account"
)

var (
	// ErrExpiredToken is returned by the client when authorization token
	// has expired. If returned, repeating the same request will acquire
	// a new authorization token.
	ErrExpiredToken = errors.New("expired auth token")

	// ErrUnauthorized is returned when the applicationKeyId and/or the
	// applicationKey are wrong.
	ErrUnauthorized = errors.New("invalid credentials")

	// timeNow is a mockable version of time.Now
	timeNow = time.Now
)

// An errorResponse contains the error caused by an API request.
type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// tokenCapability represents the capabilities of an authorization token.
type tokenCapability struct {
	BucketID     string   `json:"bucketId"`
	BucketName   string   `json:"bucketName"`
	Capabilities []string `json:"capabilities"`
	NamePrefix   string   `json:"namePrefix"`
}

// authorizationResponse is returned by the B2 API authorization call.
type authorizationResponse struct {
	// The smallest possible size of a part of a large
	// file (except the last one).
	MinimumPartSize int `json:"absoluteMinimumPartSize"`

	// The identifier for the account.
	AccountID string `json:"accountId"`

	// Contains information about what's allowed with this auth token.
	TokenCapabilities tokenCapability `json:"allowed"`

	// The base URL for all API calls except for uploading
	// and downloading files.
	APIURL string `json:"apiUrl"`

	// The base URL for all API calls using the S3 compatible API.
	S3APIURL string `json:"s3ApiUrl"`

	// The token used for all API calls that need an
	// authorization header. The token is valid for
	// at most 24 hours.
	AuthorizationToken string `json:"authorizationToken"`

	// The base URL to use for downloading files.
	DownloadURL string `json:"downloadUrl"`

	// The recommended size for each part of a large file
	// for optimal upload performance.
	RecommendedPartSize int `json:"recommendedPartSize"`
}

// Cache defines the interface for interacting with a cache.
type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
}

// Client manages communication with Backblaze API
type Client struct {
	// HTTP client used to communicate with the B2 API
	client *http.Client

	// User agent for the client
	userAgent string

	// Base URL for API requests
	baseURL *url.URL

	// Cache is used for caching authorization tokens for up to
	// 24 hours as well as various other things such as bucket
	// name to ID mappings required for many API calls
	cache Cache

	// Session holds the information obtained from the login call
	Session *Session

	// Services used for communicating with the API
	Bucket *BucketService
	File   *FileService
}

// ClientOpt are options for New.
type ClientOpt func(*Client) error

// NewClient returns a new Backblaze API client.
func NewClient(keyId, keySecret string, opts ...ClientOpt) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:    http.DefaultClient,
		Session:   &Session{},
		baseURL:   baseURL,
		userAgent: "b2/" + version.Version + " (+https://github.com/romantomjak/b2)",
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.cache == nil {
		cache, err := newDiskCache()
		if err != nil {
			return nil, err
		}
		c.cache = cache
	}

	session, err := restoreSessionFromCache(c.cache)
	if err != nil {
		return nil, fmt.Errorf("cache: %v", err)
	}

	if session.Expired() {
		session, err = c.authorizeAccount(keyId, keySecret)
		if err != nil {
			return nil, fmt.Errorf("authorize account: %v", err)
		}
		if err := commitSessionToCache(c.cache, session); err != nil {
			return nil, fmt.Errorf("cache: %v", err)
		}
	}

	c.Session = session

	// Set the new base URL after authorization
	apiURL, err := url.Parse(c.Session.APIURL)
	if err != nil {
		return nil, fmt.Errorf("parse api url: %v", err)
	}
	c.baseURL = apiURL

	c.Bucket = &BucketService{client: c}
	c.File = &FileService{client: c}

	return c, nil
}

// SetBaseURL is a client option for setting the base URL.
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

// SetCache is a client option for changing cache client.
func SetCache(cache Cache) ClientOpt {
	return func(c *Client) error {
		c.cache = cache
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
func (c *Client) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.Session.AuthorizationToken)

	return req, nil
}

// newRequest prepares a new Request.
//
// Creates a new request object without authorization data.
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	var b io.Reader
	if body != nil {
		if r, ok := body.(io.Reader); ok {
			b = r
		} else {
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(body)
			if err != nil {
				return nil, err
			}
			b = buf
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), b)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.userAgent)

	return req, nil
}

// Do sends an API request and returns the API response.
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

// authorizeAccount is used to log in to the B2 API.
func (c *Client) authorizeAccount(keyId, keySecret string) (*Session, error) {
	ctx := context.Background()

	req, err := c.newRequest(ctx, http.MethodGet, authorizationURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(keyId, keySecret)

	authResp := new(authorizationResponse)
	_, err = c.Do(req, &authResp)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		TokenExpiresAt:        timeNow().Add(time.Hour * 24),
		authorizationResponse: *authResp,
	}

	return sess, nil
}

// checkResponse checks the API response for errors and returns them if present.
//
// Any code other than 2xx is an error, and the response will contain a JSON
// error structure indicating what went wrong.
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

	if r.StatusCode == 401 {
		switch errResp.Code {
		case "expired_auth_token":
			return ErrExpiredToken
		case "unauthorized":
			return ErrUnauthorized
		}
	}

	return fmt.Errorf("%v %v %v %v: %v %v", r.Proto, r.StatusCode, r.Request.Method, r.Request.URL, errResp.Code, errResp.Message)
}

// newDiskCache creates and returns disk cache.
//
// The directory used for caching is created if it doesn't exist already
func newDiskCache() (*DiskCache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(cacheDir, "b2")

	err = os.Mkdir(path, 0700)
	if err != nil {
		// ignore "already exists" error
		if !os.IsExist(err) {
			return nil, err
		}
	}

	return NewDiskCache(path)
}
