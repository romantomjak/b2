package b2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	listFilesURL           = "b2api/v2/b2_list_file_names"
	fileUploadURL          = "b2api/v2/b2_get_upload_url"
	filePartUploadURL      = "b2api/v2/b2_get_upload_part_url"
	fileStartLargeFileURL  = "b2api/v2/b2_start_large_file"
	fileFinishLargeFileURL = "b2api/v2/b2_finish_large_file"
	fileCancelLargeFileURL = "b2api/v2/b2_cancel_large_file"
)

// File describes a File or a Folder in a Bucket
type File struct {
	AccountID       string            `json:"accountId"`
	Action          string            `json:"action"`
	BucketID        string            `json:"bucketId"`
	ContentLength   int               `json:"contentLength"`
	ContentSHA1     string            `json:"contentSha1"`
	ContentType     string            `json:"contentType"`
	FileID          string            `json:"fileId"`
	FileInfo        map[string]string `json:"fileInfo"`
	FileName        string            `json:"fileName"`
	UploadTimestamp int64             `json:"uploadTimestamp"`
}

type FilePart struct {
	Number        int64  `json:"partNumber"`
	FileID        string `json:"fileId"`
	ContentLength int64  `json:"contentLength"`
	ContentSHA1   string `json:"contentSha1"`
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

// PartUploadAuthorizationRequest represents a request to obtain a URL
// for uploading parts of a file.
type UploadPartAuthorizationRequest struct {
	FileID string `json:"fileId"`
}

// UploadAuthorization contains the information for uploading a file
// or a part of a file.
type UploadAuthorization struct {
	BucketID  string `json:"bucketId"`
	UploadURL string `json:"uploadUrl"`
	Token     string `json:"authorizationToken"`
}

// UploadRequest represents a request to upload a file.
type UploadRequest struct {
	Authorization *UploadAuthorization
	Body          io.Reader
	Key           string
	ChecksumSHA1  string
	ContentLength int64
	LastModified  time.Time
}

type UploadPartRequest struct {
	Authorization *UploadAuthorization
	PartNumber    int64
	Body          io.Reader
	ChecksumSHA1  string
	ContentLength int64
}

// StartLargeFileRequest prepares for uploading the parts of a large file.
type StartLargeFileRequest struct {
	BucketID    string            `json:"bucketId"`
	Filename    string            `json:"fileName"`
	ContentType string            `json:"contentType"`
	FileInfo    map[string]string `json:"fileInfo"`
}

// FinishLargeFileRequest converts the parts that have been uploaded into a single B2 file.
type FinishLargeFileRequest struct {
	// The ID returned by StartLargeFileRequest.
	FileID string `json:"fileId"`

	// An array of SHA1 checksums of the parts of the large file. This is used to check that
	// the parts were uploaded in the right order, and that none were missed.
	PartSHA1 []string `json:"partSha1Array"`
}

// CancelLargeFileRequest cancels the upload of a large file, and deletes all of the parts that have been uploaded.
type CancelLargeFileRequest struct {
	// The ID returned by StartLargeFileRequest.
	FileID string `json:"fileId"`
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
		return nil, nil, err
	}

	return root.Files, resp, nil
}

// Download a file
func (s *FileService) Download(ctx context.Context, url string, w io.Writer) (*http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, w)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UploadAuthorization returns the information for uploading a file.
func (s *FileService) UploadAuthorization(ctx context.Context, uploadAuthorizationRequest *UploadAuthorizationRequest) (*UploadAuthorization, *http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, fileUploadURL, uploadAuthorizationRequest)
	if err != nil {
		return nil, nil, err
	}

	auth := new(UploadAuthorization)
	resp, err := s.client.Do(req, auth)
	if err != nil {
		return nil, nil, err
	}

	return auth, resp, nil
}

// PartUploadAuthorization returns the information for uploading a part of a file.
func (s *FileService) UploadPartAuthorization(ctx context.Context, authorizationRequest *UploadPartAuthorizationRequest) (*UploadAuthorization, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, filePartUploadURL, authorizationRequest)
	if err != nil {
		return nil, err
	}

	auth := new(UploadAuthorization)
	_, err = s.client.Do(req, auth)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

// Upload a file.
func (s *FileService) Upload(ctx context.Context, uploadRequest *UploadRequest) (*File, *http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, uploadRequest.Authorization.UploadURL, uploadRequest.Body)
	if err != nil {
		return nil, nil, err
	}

	req.ContentLength = uploadRequest.ContentLength

	req.Header.Set("Authorization", uploadRequest.Authorization.Token)
	req.Header.Set("X-Bz-File-Name", url.QueryEscape(uploadRequest.Key))
	req.Header.Set("Content-Type", "b2/x-auto")
	req.Header.Set("X-Bz-Content-Sha1", uploadRequest.ChecksumSHA1)
	req.Header.Set("X-Bz-Info-src_last_modified_millis", fmt.Sprintf("%d", uploadRequest.LastModified.Unix()*1000))

	file := new(File)
	resp, err := s.client.Do(req, file)
	if err != nil {
		return nil, nil, err
	}

	return file, resp, nil
}

func (s *FileService) UploadPart(ctx context.Context, uploadRequest *UploadPartRequest) (*FilePart, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, uploadRequest.Authorization.UploadURL, uploadRequest.Body)
	if err != nil {
		return nil, err
	}

	req.ContentLength = uploadRequest.ContentLength

	req.Header.Set("Authorization", uploadRequest.Authorization.Token)
	req.Header.Set("X-Bz-Part-Number", fmt.Sprintf("%d", uploadRequest.PartNumber))
	req.Header.Set("Content-Type", "b2/x-auto")
	req.Header.Set("X-Bz-Content-Sha1", uploadRequest.ChecksumSHA1)

	part := new(FilePart)
	_, err = s.client.Do(req, part)
	if err != nil {
		return nil, err
	}

	return part, nil
}

func (s *FileService) StartLargeFile(ctx context.Context, uploadRequest *StartLargeFileRequest) (*File, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, fileStartLargeFileURL, uploadRequest)
	if err != nil {
		return nil, err
	}

	file := new(File)
	_, err = s.client.Do(req, file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) FinishLargeFile(ctx context.Context, uploadRequest *FinishLargeFileRequest) (*File, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, fileFinishLargeFileURL, uploadRequest)
	if err != nil {
		return nil, err
	}

	file := new(File)
	_, err = s.client.Do(req, file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) CancelLargeFile(ctx context.Context, uploadRequest *CancelLargeFileRequest) (*File, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, fileCancelLargeFileURL, uploadRequest)
	if err != nil {
		return nil, err
	}

	file := new(File)
	_, err = s.client.Do(req, file)
	if err != nil {
		return nil, err
	}

	return file, nil
}
