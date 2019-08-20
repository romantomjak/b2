package b2

import (
	"net/http"
	"os"
)

const (
	authorizationURL = "b2api/v2/b2_authorize_account"
)

// TokenCapability represents the capabilities of a token
type TokenCapability struct {
	BucketID     string   `json:"bucketId"`
	BucketName   string   `json:"bucketName"`
	Capabilities []string `json:"capabilities"`
	NamePrefix   string   `json:"namePrefix"`
}

// Authorization represents the authorization response from the B2 API
type Authorization struct {
	AbsoluteMinimumPartSize int             `json:"absoluteMinimumPartSize"`
	AccountID               string          `json:"accountId"`
	Allowed                 TokenCapability `json:"allowed"`
	APIURL                  string          `json:"apiUrl"`
	AuthorizationToken      string          `json:"authorizationToken"`
	DownloadURL             string          `json:"downloadUrl"`
	RecommendedPartSize     int             `json:"recommendedPartSize"`
}

// authorize is used to log in to the B2 API
//
// Authorization API call returns a token and a URL that should be used as
// the base URL for subsequent API calls
func (c *Client) authorize() (*Authorization, error) {
	req, err := c.newRequest(http.MethodGet, authorizationURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(os.Getenv("B2_KEY_ID"), os.Getenv("B2_KEY_SECRET"))

	auth := new(Authorization)
	_, err = c.Do(req, &auth)
	if err != nil {
		return nil, err
	}

	return auth, nil
}