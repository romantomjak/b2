package b2

import (
	"context"
	"net/http"
	"time"
)

const (
	authorizeAccountURL   = "b2api/v2/b2_authorize_account"
	authorizationCacheKey = "authorization"
)

// AccountAuthorization is returned by the B2 API authorization call.
type AccountAuthorization struct {
	// The identifier for the account.
	AccountID string `json:"accountId"`

	// The token used for all API calls that need an authorization header.
	// The token is valid for at most 24 hours.
	AuthorizationToken string `json:"authorizationToken"`

	// Contains information about what's allowed with this auth token.
	TokenCapabilities TokenCapability `json:"allowed"`

	// The base URL for all API calls except for uploading and downloading
	// files.
	APIURL string `json:"apiUrl"`

	// The base URL to use for downloading files.
	DownloadURL string `json:"downloadUrl"`

	// The recommended size for each part of a large file
	// for optimal upload performance.
	RecommendedPartSize int `json:"recommendedPartSize"`

	// The smallest possible size of a part of a large
	// file (except the last one).
	AbsoluteMinimumPartSize int `json:"absoluteMinimumPartSize"`

	// The base URL for all API calls using the S3 compatible API.
	S3APIURL string `json:"s3ApiUrl"`
}

// TokenCapability represents the capabilities of an authorization token.
type TokenCapability struct {
	// BucketID is set when access is restricted to a single bucket.
	BucketID string `json:"bucketId"`

	// BucketName is the name of the bucket identified by BucketID. It's possible
	// that BucketID is set to a bucket that no longer exists, in which case this
	// field will be empty. It's also empty when BucketID is empty.
	BucketName string `json:"bucketName"`

	// A list of strings, each one naming a capability the key has.
	Capabilities []string `json:"capabilities"`

	// NamePrefix is set when access is restricted to files whose names start with
	// the prefix.
	NamePrefix string `json:"namePrefix"`
}

// accountAuthorizationWithExpiryTimestamp is used for keeping track of when
// the authorization token expires.
type accountAuthorizationWithExpiryTimestamp struct {
	AccountAuthorization

	// Timestamp of when the token will become invalid.
	TokenExpiresAt time.Time `json:"tokenExpiresAt"`
}

// Expired returns whether the authorization token has expired.
func (a *accountAuthorizationWithExpiryTimestamp) Expired() bool {
	return timeNow().After(a.TokenExpiresAt)
}

type AccountAuthorizeRequest struct {
	// KeyID is the ID of the key.
	KeyID string

	// KeySecret is the secret part of the key.
	KeySecret string
}

// AuthorizationService handles communication with the Authorization related
// methods of the B2 API.
type AuthorizationService struct {
	client *Client
}

// AuthorizeAccount is used to log in to the B2 API.
func (s *AuthorizationService) AuthorizeAccount(ctx context.Context, authorizationRequest *AccountAuthorizeRequest) (*AccountAuthorization, error) {
	// Have we already authorized the account?
	cachedAuth := new(accountAuthorizationWithExpiryTimestamp)
	if err := s.client.cache.Get(authorizationCacheKey, cachedAuth); err != nil {
		return nil, err
	}

	if !cachedAuth.Expired() {
		return &cachedAuth.AccountAuthorization, nil
	}

	// Otherwise obtain new authorization data.
	req, err := s.client.newRequest(ctx, http.MethodGet, authorizeAccountURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(authorizationRequest.KeyID, authorizationRequest.KeySecret)

	auth := new(AccountAuthorization)
	if _, err := s.client.Do(req, auth); err != nil {
		return nil, err
	}

	// Cache the new authorization data.
	cachedAuth = &accountAuthorizationWithExpiryTimestamp{
		AccountAuthorization: *auth,
		TokenExpiresAt:       timeNow().Add(time.Hour * 24),
	}
	if err := s.client.cache.Set(authorizationCacheKey, cachedAuth); err != nil {
		return nil, err
	}

	return auth, err
}
