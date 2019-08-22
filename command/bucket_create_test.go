package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestCreateBucketCommand_RequiresBucketName(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <bucket-name>")
}

func TestCreateBucketCommand_RequiresValidBucketType(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"-type=foo", "my-bucket"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, `-type must be either "public" or "private"`)
}

func TestCreateBucketCommand_BucketCreateRequest(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_create_bucket", func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertHttpMethod(t, r.Method, "POST")

		var got map[string]string
		json.NewDecoder(r.Body).Decode(&got)

		testutil.AssertEqual(t, got["accountId"], "abc123")
		testutil.AssertEqual(t, got["bucketName"], "my-bucket")
		testutil.AssertEqual(t, got["bucketType"], "allPrivate")
	})

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	cmd.Run([]string{"my-bucket"})
}

func TestCreateBucketCommand_PrintsCreatedBucketID(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_create_bucket", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"accountId" : "010203040506",
			"bucketId" : "4a48fe8875c6214145260818",
			"bucketInfo" : {},
			"bucketName" : "my-bucket",
			"bucketType" : "allPrivate",
			"lifecycleRules" : []
		}`)
	})

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"my-bucket"})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, fmt.Sprintf("Bucket %q created with ID %q", "my-bucket", "4a48fe8875c6214145260818"))
}
