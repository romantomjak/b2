package command

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPutCommand_FilePrefix(t *testing.T) {
	filenameTests := []struct {
		originalFilename    string
		destinationFilename string
		filePrefix          string
	}{
		{"file1.txt", "bucket", "file1.txt"},
		{"file1.txt", "bucket/", "file1.txt"},
		{"file1.txt", "bucket/dir", "dir"},
		{"file1.txt", "bucket/dir/", "dir/file1.txt"},
		{"file1.txt", "bucket/dir/a", "dir/a"},
		{"file1.txt", "bucket/dir/a/", "dir/a/file1.txt"},
	}
	for _, tt := range filenameTests {
		t.Run(tt.destinationFilename, func(t *testing.T) {
			_, filePrefix := destinationBucketAndFilename(tt.originalFilename, tt.destinationFilename)
			assert.Equal(t, tt.filePrefix, filePrefix)
		})
	}
}

func TestPutCommand_CanUploadFile(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_list_buckets", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"buckets": [
			{
				"accountId": "30f20426f0b1",
				"bucketId": "87ba238875c6214145260818",
				"bucketInfo": {},
				"bucketName": "Secret-Documents",
				"bucketType": "allPrivate",
				"lifecycleRules": []
			} ]
		}`)
	})

	mux.HandleFunc("/b2api/v2/b2_get_upload_url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"bucketId": "87ba238875c6214145260818",
			"uploadUrl": "%s",
			"authorizationToken": "some-secret-token"
		}`, server.URL)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileBytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		assert.NoError(t, err)
		assert.Equal(t, []byte("This file is not empty."), fileBytes)

		fmt.Fprint(w, `{
			"fileId": "4_h4a48fe8875c6214145260818_f000000000000472a_d20140104_m032022_c001_v0000123_t0104",
			"fileName": "typing_test.txt",
			"accountId": "d522aa47a10f",
			"bucketId": "4a48fe8875c6214145260818",
			"contentLength": 46,
			"contentSha1": "bae5ed658ab3546aee12f23f36392f35dba1ebdd",
			"contentType": "text/plain",
			"fileInfo": {
				"author": "unknown"
			}
		}`)
	})

	cache, _ := b2.NewInMemoryCache()
	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &PutCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	tmpFile, _ := ioutil.TempFile(os.TempDir(), "b2-cli-test-")
	defer os.Remove(tmpFile.Name())

	tmpFile.Write([]byte("This file is not empty."))
	tmpFile.Close()

	src := tmpFile.Name()
	dst := "Secret-Documents"

	code := cmd.Run([]string{src, dst})
	assert.Equal(t, 0, code)

	out := ui.OutputWriter.String()
	filename := fmt.Sprintf("%s/%s", dst, path.Base(tmpFile.Name()))
	assert.Contains(t, out, fmt.Sprintf("Uploaded %q to %q", src, filename))
}
