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
	testutil.AssertContains(t, out, "Kitten-Videos")
	testutil.AssertContains(t, out, "Puppy-Videos")
	testutil.AssertContains(t, out, "Vacation-Pictures")
}
