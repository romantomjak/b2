package command

import (
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestListCommand_AcceptsPathArgument(t *testing.T) {
	client := b2.NewClient(nil)
	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"one", "two"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <path>")
}

func TestListCommand_CanListBuckets(t *testing.T) {
	body := `{
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
	}`
	client := b2.NewClient(&testutil.FakeHTTPClient{
		Response: testutil.HTTPResponse(http.StatusOK, body),
	})
	client.Token = "TEST"

	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, "Kitten-Videos")
	testutil.AssertContains(t, out, "Puppy-Videos")
	testutil.AssertContains(t, out, "Vacation-Pictures")
}
