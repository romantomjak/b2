package b2

import (
	"net/http"
	"net/url"
	"os"
)

const (
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

// authorizeAccount is used to log in to the B2 API
//
// This must be the very first API call to obtain essential account information
func (c *Client) authorizeAccount() (*authorizeAccount, error) {
	req, err := c.newRequest(http.MethodGet, authorizeAccountURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(os.Getenv("B2_KEY_ID"), os.Getenv("B2_KEY_SECRET"))

	account := new(authorizeAccount)
	_, err = c.Do(req, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
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
