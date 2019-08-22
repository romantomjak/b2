package command

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestListCommand_AcceptsPathArgument(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))
	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"one", "two"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <path>")
}

func TestListCommand_ListFilesInBucket(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_list_buckets", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"buckets": [
			{
				"accountId": "abc123",
				"bucketId": "4a48fe8875c6214145260818",
				"bucketInfo": {},
				"bucketName" : "my-bucket",
				"bucketType": "allPrivate",
				"lifecycleRules": []
			} ]
		}`)
	})

	mux.HandleFunc("/b2api/v2/b2_list_file_names", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"files": [
			{
				"accountId": "abc123",
				"action": "upload",
				"bucketId": "4a48fe8875c6214145260818",
				"contentLength": 7,
				"contentSha1": "dc724af18fbdd4e59189f5fe768a5f8311527050",
				"contentType": "text/plain",
				"fileId": "4_zb2f6f21365e1d29f6c580f18_f10904e5ca06493a1_d20180914_m223119_c002_v0001094_t0002",
				"fileInfo": {
					"src_last_modified_millis": "1536964184056"
				},
				"fileName": "testing.txt",
				"uploadTimestamp": 1536964279000
			},
			{
				"accountId": "abc123",
				"action": "upload",
				"bucketId": "4a48fe8875c6214145260818",
				"contentLength": 8,
				"contentSha1": "596b29ec9afea9e461a20610d150939b9c399d93",
				"contentType": "text/plain",
				"fileId": "4_zb2f6f21365e1d29f6c580f18_f10076875fe98d4af_d20180914_m223128_c002_v0001108_t0050",
				"fileInfo": {
					"src_last_modified_millis": "1536964200750"
				},
				"fileName": "testing2.txt",
				"uploadTimestamp": 1536964288000
			}
			],
			"nextFileName": null
		}`)
	})

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"my-bucket"})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, "testing.txt")
	testutil.AssertContains(t, out, "testing2.txt")
}
