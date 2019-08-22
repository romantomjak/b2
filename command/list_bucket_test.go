package command

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
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

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, "Kitten-Videos/")
	testutil.AssertContains(t, out, "Puppy-Videos/")
	testutil.AssertContains(t, out, "Vacation-Pictures/")
}

func TestListCommand_LookupBucketByName(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_list_buckets", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"buckets": []}`)
	})

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

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
			testutil.AssertEqual(t, code, 1)

			out := ui.ErrorWriter.String()
			testutil.AssertContains(t, out, `bucket with name "bucket-name" was not found`)

			ui.ErrorWriter.Reset()
		})
	}
}
