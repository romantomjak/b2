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

	// timeNow is a mockable version of time.Now
	timeNow = time.Now
)

// An errorResponse contains the error caused by an API request
type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// tokenCapability represents the capabilities of an authorization token
type tokenCapability struct {
	BucketID     string   `json:"bucketId"`
	BucketName   string   `json:"bucketName"`
	Capabilities []string `json:"capabilities"`
	NamePrefix   string   `json:"namePrefix"`
}

// authorizationResponse is returned by the B2 API authorization call
type authorizationResponse struct {
	AbsoluteMinimumPartSize int64           `json:"absoluteMinimumPartSize"`
	AccountID               string          `json:"accountId"`
	Allowed                 tokenCapability `json:"allowed"`
	APIURL                  string          `json:"apiUrl"`
	AuthorizationToken      string          `json:"authorizationToken"`
	DownloadURL             string          `json:"downloadUrl"`
	RecommendedPartSize     int64           `json:"recommendedPartSize"`
}

type authorization struct {
	authorizationResponse
	TokenExpiresAt time.Time `json:"tokenExpiresAt"`
}

// Cache defines the interface for interacting with a cache
type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
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

	// The client for accessing cache
	cache Cache

	// The account identifier
	AccountID string

	// The base URL for downloading files
	DownloadURL *url.URL

	// The recommended size for each part of a large file
	RecommendedPartSize int64

	// Services used for communicating with the API
	Bucket *BucketService
	File   *FileService
}

// ClientOpt are options for New
type ClientOpt func(*Client) error

// NewClient returns a new Backblaze API client
func NewClient(keyId, keySecret string, opts ...ClientOpt) (*Client, error) {
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

	if c.cache == nil {
		cache, err := newDiskCache()
		if err != nil {
			return nil, err
		}
		c.cache = cache
	}

	// attempts := 0
	// for {
	// 	attempts++
	// 	authz, err := s.authorize(keyId, keySecret)
	// 	if err != nil {
	// 		if err == ErrExpiredToken && attempts < 2 {
	// 			continue
	// 		}
	// 		return nil, err
	// 	}
	// 	return authz, nil
	// }

	err := c.authorize(keyId, keySecret)
	if err != nil {
		return nil, fmt.Errorf("authorization: %v", err)
	}

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

// SetCache is a client option for changing cache client
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

	req.Header.Add("Authorization", c.auth.AuthorizationToken)

	return req, nil
}

// newRequest prepares a new Request
//
// Creates a new request object without authorization data
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

// authorize is used to log in to the B2 API
//
// Authorization API call returns a token and a URL that should be used as
// the base URL for subsequent API calls
func (c *Client) authorize(keyId, keySecret string) error {
	var auth *authorization

	// use cached authorization or request a fresh token
	auth, err := authorizationFromCache(c.cache)
	if err != nil {
		ctx := context.Background()

		req, err := c.newRequest(ctx, http.MethodGet, authorizationURL, nil)
		if err != nil {
			return err
		}

		req.SetBasicAuth(keyId, keySecret)

		authResp := new(authorizationResponse)
		_, err = c.Do(req, &authResp)
		if err != nil {
			return err
		}

		auth = &authorization{
			TokenExpiresAt:        timeNow().Add(time.Hour * 24),
			authorizationResponse: *authResp,
		}

		cacheAuthorization(auth, c.cache)
		if err != nil {
			// TODO: write to log
			fmt.Println(err)
		}
	}

	apiURL, err := url.Parse(auth.APIURL)
	if err != nil {
		return err
	}

	downloadURL, err := url.Parse(auth.DownloadURL)
	if err != nil {
		return err
	}

	// TODO: synchronize access to c.auth

	c.baseURL = apiURL
	c.auth = auth
	c.AccountID = auth.AccountID
	c.DownloadURL = downloadURL
	c.RecommendedPartSize = auth.RecommendedPartSize

	return nil
}

func authorizationFromCache(cache Cache) (*authorization, error) {
	val, err := cache.Get("authorization")
	if err != nil {
		return nil, err
	}

	auth, ok := val.(authorization)
	if !ok {
		return nil, fmt.Errorf("cannot cast %T as authorization", val)
	}

	if timeNow().After(auth.TokenExpiresAt) {
		return nil, ErrExpiredToken
	}

	return &auth, nil
}

func cacheAuthorization(auth *authorization, cache Cache) error {
	return cache.Set("authorization", auth)
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

// newDiskCache creates and returns disk cache
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
		if os.IsExist(err) == false {
			return nil, err
		}
	}

	return NewDiskCache(path)
}
