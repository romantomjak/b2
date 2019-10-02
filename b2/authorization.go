package b2

import (
	"net/http"
	"net/url"
	"os"
)

const (
	authorizationURL = "b2api/v2/b2_authorize_account"
)

// tokenCapability represents the capabilities of a token
type tokenCapability struct {
	BucketID     string   `json:"bucketId"`
	BucketName   string   `json:"bucketName"`
	Capabilities []string `json:"capabilities"`
	NamePrefix   string   `json:"namePrefix"`
}

// authorization represents the authorization response from the B2 API
type authorization struct {
	AbsoluteMinimumPartSize int             `json:"absoluteMinimumPartSize"`
	AccountID               string          `json:"accountId"`
	Allowed                 tokenCapability `json:"allowed"`
	APIURL                  string          `json:"apiUrl"`
	AuthorizationToken      string          `json:"authorizationToken"`
	DownloadURL             string          `json:"downloadUrl"`
	RecommendedPartSize     int             `json:"recommendedPartSize"`
}

// authorize is used to log in to the B2 API
//
// Authorization API call returns a token and a URL that should be used as
// the base URL for subsequent API calls
func (c *Client) authorize() error {
	req, err := c.newRequest(http.MethodGet, authorizationURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(os.Getenv("B2_KEY_ID"), os.Getenv("B2_KEY_SECRET"))

	auth := new(authorization)
	_, err = c.Do(req, &auth)
	if err != nil {
		return err
	}

	apiURL, err := url.Parse(auth.APIURL)
	if err != nil {
		return err
	}

	downloadURL, err := url.Parse(auth.DownloadURL)
	if err != nil {
		return err
	}

	c.auth = auth
	c.baseURL = apiURL
	c.DownloadURL = downloadURL

	return nil
}
