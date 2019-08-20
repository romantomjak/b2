package command

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestListCommand_AcceptsPathArgument(t *testing.T) {
	client, _ := b2.NewClient(b2.SetAuthentication(&b2.Authorization{
		AbsoluteMinimumPartSize: 5000000,
		AccountID:               "abc123",
		Allowed: b2.TokenCapability{
			BucketID:     "my-bucket",
			BucketName:   "MY BUCKET",
			Capabilities: []string{"listBuckets", "listFiles", "readFiles", "shareFiles", "writeFiles", "deleteFiles"},
			NamePrefix:   "",
		},
		APIURL:              "https://api123.backblazeb2.com",
		AuthorizationToken:  "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=",
		DownloadURL:         "https://f123.backblazeb2.com",
		RecommendedPartSize: 100000000,
	}))
	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"one", "two"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <path>")
}

func TestListCommand_CanListBuckets(t *testing.T) {
	mux := http.NewServeMux()
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
	server := httptest.NewServer(mux)
	client, _ := b2.NewClient(b2.SetAuthentication(&b2.Authorization{
		AbsoluteMinimumPartSize: 5000000,
		AccountID:               "abc123",
		Allowed: b2.TokenCapability{
			BucketID:     "my-bucket",
			BucketName:   "MY BUCKET",
			Capabilities: []string{"listBuckets", "listFiles", "readFiles", "shareFiles", "writeFiles", "deleteFiles"},
			NamePrefix:   "",
		},
		APIURL:              server.URL,
		AuthorizationToken:  "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=",
		DownloadURL:         "https://f123.backblazeb2.com",
		RecommendedPartSize: 100000000,
	}))

	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, "Kitten-Videos")
	testutil.AssertContains(t, out, "Puppy-Videos")
	testutil.AssertContains(t, out, "Vacation-Pictures")
}
