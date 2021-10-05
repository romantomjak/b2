package b2

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

const (
	listFilesURL  = "b2api/v2/b2_list_file_names"
	fileUploadURL = "b2api/v2/b2_get_upload_url"
)

// File describes a File or a Folder in a Bucket
type File struct {
	AccountID       string            `json:"accountId"`
	Action          string            `json:"action"`
	BucketID        string            `json:"bucketId"`
	ContentLength   int               `json:"contentLength"`
	ContentSha1     string            `json:"contentSha1"`
	ContentType     string            `json:"contentType"`
	FileID          string            `json:"fileId"`
	FileInfo        map[string]string `json:"fileInfo"`
	FileName        string            `json:"fileName"`
	UploadTimestamp int64             `json:"uploadTimestamp"`
}

// FileListRequest represents a request to list files in a Bucket
type FileListRequest struct {
	BucketID      string `json:"bucketId"`
	StartFileName string `json:"startFileName,omitempty"`
	MaxFileCount  int    `json:"maxFileCount,omitempty"`
	Prefix        string `json:"prefix,omitempty"`
	Delimiter     string `json:"delimiter,omitempty"`
}

type fileListRoot struct {
	Files        []File `json:"files"`
	NextFileName string `json:"nextFileName"`
}

// UploadAuthorizationRequest represents a request to obtain a URL for uploading files
type UploadAuthorizationRequest struct {
	BucketID string `json:"bucketId"`
}

// UploadAuthorization contains the information for uploading a file
type UploadAuthorization struct {
	BucketID  string `json:"bucketId"`
	UploadURL string `json:"uploadUrl"`
	Token     string `json:"authorizationToken"`
}

// FileService handles communication with the File related methods of the
// B2 API
type FileService struct {
	client *Client
}

// List files in a Bucket
func (s *FileService) List(ctx context.Context, listRequest *FileListRequest) ([]File, *http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, listFilesURL, listRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(fileListRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Files, resp, err
}

// Download a file
func (s *FileService) Download(ctx context.Context, url string, w io.Writer) (*http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, w)
	if err != nil {
		return resp, err
	}

	return resp, err
}

// UploadAuthorization returns the information for uploading a file
func (s *FileService) UploadAuthorization(ctx context.Context, uploadAuthorizationRequest *UploadAuthorizationRequest) (*UploadAuthorization, *http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, fileUploadURL, uploadAuthorizationRequest)
	if err != nil {
		return nil, nil, err
	}

	auth := new(UploadAuthorization)
	resp, err := s.client.Do(req, auth)
	if err != nil {
		return nil, resp, err
	}

	return auth, resp, nil
}

type UploadInput struct {
	Authorization *UploadAuthorization
	Body          io.Reader
	Key           string
	ContentSHA1   string
	ContentLength int64
	Metadata      map[string]string
}

// Upload a file
func (s *FileService) Upload(ctx context.Context, input *UploadInput) (*File, *http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, input.Authorization.UploadURL, input.Body)
	if err != nil {
		return nil, nil, err
	}

	req.ContentLength = input.ContentLength

	req.Header.Set("Authorization", input.Authorization.Token)
	req.Header.Set("X-Bz-File-Name", url.QueryEscape(input.Key))
	req.Header.Set("Content-Type", "b2/x-auto")
	req.Header.Set("X-Bz-Content-Sha1", input.ContentSHA1)

	for key, val := range input.Metadata {
		req.Header.Set("X-Bz-Info-"+key, val)
	}

	file := new(File)
	resp, err := s.client.Do(req, file)
	if err != nil {
		return nil, resp, err
	}

	return file, resp, nil
}
