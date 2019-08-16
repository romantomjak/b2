package command

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
)

func TestListCommand_AcceptsPathArgument(t *testing.T) {
	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"one", "two"})
	assertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	assertContains(t, out, "This command takes one argument: <path>")
}

func TestListCommand_CanListBuckets(t *testing.T) {
	setup()
	defer teardown()

	// manually set the token so the client does not perform authentication
	// FIXME: maybe extract into a performTestAuthentication()?
	client.Token = "TEST"

	mux.HandleFunc("/b2api/v2/b2_list_buckets", func(w http.ResponseWriter, r *http.Request) {
		assertHttpMethod(t, r.Method, "POST")

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

	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	assertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	assertContains(t, out, "Kitten-Videos")
	assertContains(t, out, "Puppy-Videos")
	assertContains(t, out, "Vacation-Pictures")
}
