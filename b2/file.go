package b2

import (
	"io"
	"net/http"
)

const (
	listFilesURL = "b2api/v2/b2_list_file_names"
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

// FileService handles communication with the File related methods of the
// B2 API
type FileService struct {
	client *Client
}

// List files in a Bucket
func (s *FileService) List(listRequest *FileListRequest) ([]File, *http.Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, listFilesURL, listRequest)
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
func (s *FileService) Download(url string, w io.Writer) (*http.Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, w)
	if err != nil {
		return resp, err
	}

	return resp, err
}
