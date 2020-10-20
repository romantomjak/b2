package b2

import (
	"net/http"
)

const (
	createBucketURL = "b2api/v2/b2_create_bucket"
	listBucketsURL  = "b2api/v2/b2_list_buckets"
)

// Bucket is used to represent a B2 Bucket
type Bucket struct {
	AccountID      string                `json:"accountId"`
	ID             string                `json:"bucketId"`
	Info           map[string]string     `json:"bucketInfo"`
	Name           string                `json:"bucketName"`
	Type           string                `json:"bucketType"`
	LifecycleRules []BucketLifecycleRule `json:"lifecycleRules"`
	Revision       int                   `json:"revision"`
}

// BucketCreateRequest represents a request to create a Bucket
type BucketCreateRequest struct {
	AccountID      string                `json:"accountId"`
	Name           string                `json:"bucketName"`
	Type           string                `json:"bucketType"`
	Info           map[string]string     `json:"bucketInfo,omitempty"`
	CorsRules      []BucketCorsRule      `json:"corsRules,omitempty"`
	LifecycleRules []BucketLifecycleRule `json:"lifecycleRules,omitempty"`
}

// BucketCorsRule is used to represent a Bucket's CORS rule
//
// See more on https://www.backblaze.com/b2/docs/cors_rules.html
type BucketCorsRule struct {
	Name              string   `json:"corsRuleName"`
	AllowedOrigins    []string `json:"allowedOrigins"`
	AllowedHeaders    []string `json:"allowedHeaders"`
	AllowedOperations []string `json:"allowedOperations"`
	ExposeHeaders     []string `json:"exposeHeaders"`
	MaxAgeSeconds     int      `json:"maxAgeSeconds"`
}

// BucketLifecycleRule tells B2 to automatically hide and/or delete old files
//
// See more on https://www.backblaze.com/b2/docs/lifecycle_rules.html
type BucketLifecycleRule struct {
	DaysFromHidingToDeleting  int    `json:"daysFromHidingToDeleting"`
	DaysFromUploadingToHiding int    `json:"daysFromUploadingToHiding"`
	FileNamePrefix            string `json:"fileNamePrefix"`
}

// BucketListRequest represents a request to list Buckets
type BucketListRequest struct {
	AccountID string `json:"accountId"`
	BucketID  string `json:"bucketId,omitempty"`
	Name      string `json:"bucketName,omitempty"`
	Types     string `json:"bucketTypes,omitempty"`
}

type bucketListRoot struct {
	Buckets []Bucket `json:"buckets"`
}

// BucketService handles communication with the Bucket related methods of the
// B2 API
type BucketService struct {
	client *Client
}

// Create a new Bucket
func (s *BucketService) Create(createRequest *BucketCreateRequest) (*Bucket, *http.Response, error) {
	// TODO: check authorization for creating buckets

	req, err := s.client.NewRequest(http.MethodPost, createBucketURL, createRequest)
	if err != nil {
		return nil, nil, err
	}

	bucket := new(Bucket)
	resp, err := s.client.Do(req, bucket)
	if err != nil {
		return nil, resp, err
	}

	return bucket, resp, err
}

// List all Buckets
func (s *BucketService) List(listRequest *BucketListRequest) ([]Bucket, *http.Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, listBucketsURL, listRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(bucketListRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Buckets, resp, err
}
