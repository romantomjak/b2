package b2

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	listFilesURL           = "b2api/v2/b2_list_file_names"
	fileUploadURL          = "b2api/v2/b2_get_upload_url"
	filePartUploadURL      = "b2api/v2/b2_get_upload_part_url"
	fileStartLargeFileURL  = "b2api/v2/b2_start_large_file"
	fileFinishLargeFileURL = "b2api/v2/b2_finish_large_file"
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

// PartUploadAuthorizationRequest represents a request to obtain a URL
// for uploading parts of a file.
type PartUploadAuthorizationRequest struct {
	FileID string `json:"fileId"`
}

// UploadAuthorization contains the information for uploading a file
// or a part of a file.
type UploadAuthorization struct {
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

	return s.uploadAuthorization(req)
}

// PartUploadAuthorization returns the information for uploading a part of a file
func (s *FileService) PartUploadAuthorization(ctx context.Context, uploadAuthorizationRequest *PartUploadAuthorizationRequest) (*UploadAuthorization, *http.Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, filePartUploadURL, uploadAuthorizationRequest)
	if err != nil {
		return nil, nil, err
	}

	return s.uploadAuthorization(req)
}

func (s *FileService) uploadAuthorization(req *http.Request) (*UploadAuthorization, *http.Response, error) {
	auth := new(UploadAuthorization)
	resp, err := s.client.Do(req, auth)
	if err != nil {
		return nil, resp, err
	}
	return auth, resp, nil
}

// Upload a file
func (s *FileService) Upload(ctx context.Context, uploadAuthorization *UploadAuthorization, src, dst string) (*File, *http.Response, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	hash := sha1.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return nil, nil, err
	}
	sha1 := fmt.Sprintf("%x", hash.Sum(nil))

	f.Seek(0, 0)

	req, err := s.client.NewRequest(ctx, http.MethodPost, uploadAuthorization.UploadURL, f)
	if err != nil {
		return nil, nil, err
	}

	req.ContentLength = info.Size()

	req.Header.Set("Authorization", uploadAuthorization.Token)
	req.Header.Set("X-Bz-File-Name", url.QueryEscape(dst))
	req.Header.Set("Content-Type", "b2/x-auto")
	req.Header.Set("X-Bz-Content-Sha1", sha1)
	req.Header.Set("X-Bz-Info-src_last_modified_millis", fmt.Sprintf("%d", info.ModTime().Unix()*1000))

	file := new(File)
	resp, err := s.client.Do(req, file)
	if err != nil {
		return nil, resp, err
	}

	return file, resp, nil
}

type UploadRequest struct {
	Body          io.Reader
	ContentSHA1   string
	Authorization *UploadAuthorization
	PartNumber    int64
}

type FilePart struct {
	// Number is the ID of the part
	Number int64 `json:"partNumber"`

	// FileID is the ID of the File this part is part of.
	FileID string `json:"fileId"`

	// Size returns the number of bytes stored in the part.
	Size int64 `json:"contentLength"`

	// SHA1 is the SHA1 of the bytes stored in the part.
	SHA1 string `json:"contentSha1"`
}

// Upload a part of a file
func (s *FileService) UploadPart(ctx context.Context, uploadRequest *UploadRequest) (*FilePart, *http.Response, error) {
	// TODO: the authorization should really be obtained here, but
	// 	   : backblaze requires that each thread gets it's own auth

	req, err := s.client.NewRequest(ctx, http.MethodPost, uploadRequest.Authorization.UploadURL, uploadRequest.Body)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", uploadRequest.Authorization.Token)
	req.Header.Set("Content-Type", "b2/x-auto")
	req.Header.Set("X-Bz-Content-Sha1", uploadRequest.ContentSHA1)

	part := new(FilePart)
	resp, err := s.client.Do(req, part)
	if err != nil {
		return nil, resp, err
	}

	return part, resp, nil
}

type LargeFileRequest struct {
	BucketID    string            `json:"bucketId"`
	Filename    string            `json:"fileName"`
	ContentType string            `json:"contentType"`
	FileInfo    map[string]string `json:"fileInfo"`
}

type LargeFileResponse struct {
	FileID      string            `json:"fileId"`
	BucketID    string            `json:"bucketId"`
	Filename    string            `json:"fileName"`
	ContentType string            `json:"contentType"`
	FileInfo    map[string]string `json:"fileInfo"`
}

// Upload a part of a file
func (s *FileService) StartLargeFile(ctx context.Context, largeFileRequest *LargeFileRequest) (*LargeFileResponse, *http.Response, error) {
	// TODO: the authorization should really be obtained here, but
	// 	   : backblaze requires that each thread gets it's own auth

	req, err := s.client.NewRequest(ctx, http.MethodPost, fileStartLargeFileURL, largeFileRequest)
	if err != nil {
		return nil, nil, err
	}

	file := new(LargeFileResponse)
	resp, err := s.client.Do(req, file)
	if err != nil {
		return nil, resp, err
	}

	return file, resp, nil
}

type FinishLargeFileRequest struct {
	FileID    string   `json:"fileId"`
	PartSHA1s []string `json:"partSha1Array"`
}

// Upload a part of a file
func (s *FileService) FinishLargeFile(ctx context.Context, largeFileRequest *FinishLargeFileRequest) (*LargeFileResponse, *http.Response, error) {
	// TODO: the authorization should really be obtained here, but
	// 	   : backblaze requires that each thread gets it's own auth

	req, err := s.client.NewRequest(ctx, http.MethodPost, fileFinishLargeFileURL, largeFileRequest)
	if err != nil {
		return nil, nil, err
	}

	file := new(LargeFileResponse)
	resp, err := s.client.Do(req, file)
	if err != nil {
		return nil, resp, err
	}

	return file, resp, nil
}
