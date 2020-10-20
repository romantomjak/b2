package command

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestListCommand_CanListBuckets(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_list_buckets", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"buckets": [
			{
				"accountId": "30f20426f0b1",
				"bucketId": "4a48fe8875c6214145260818",
				"bucketInfo": {},
				"bucketName" : "Kitten-Videos",
				"bucketType": "allPrivate",
				"lifecycleRules": []
			},
			{
				"accountId": "30f20426f0b1",
				"bucketId" : "5b232e8875c6214145260818",
				"bucketInfo": {},
				"bucketName": "Puppy-Videos",
				"bucketType": "allPublic",
				"lifecycleRules": []
			},
			{
				"accountId": "30f20426f0b1",
				"bucketId": "87ba238875c6214145260818",
				"bucketInfo": {},
				"bucketName": "Vacation-Pictures",
				"bucketType" : "allPrivate",
				"lifecycleRules": []
			} ]
		}`)
	})

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &ListCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	code := cmd.Run([]string{})
	assert.Equal(t, 0, code)

	out := ui.OutputWriter.String()
	assert.Contains(t, out, "Kitten-Videos/")
	assert.Contains(t, out, "Puppy-Videos/")
	assert.Contains(t, out, "Vacation-Pictures/")
}

func TestListCommand_LookupBucketByName(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_list_buckets", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"buckets": []}`)
	})

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &ListCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	tc := []struct {
		bucketName string
	}{
		{"bucket-name"},
		{"bucket-name/help"},
		{"bucket-name/help/myfile"},
	}
	for _, tt := range tc {
		t.Run(tt.bucketName, func(t *testing.T) {
			code := cmd.Run([]string{tt.bucketName})
			assert.Equal(t, 1, code)

			out := ui.ErrorWriter.String()
			assert.Contains(t, out, `bucket with name "bucket-name" was not found`)

			ui.ErrorWriter.Reset()
		})
	}
}
