package command

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestCreateBucketCommand_RequiresBucketName(t *testing.T) {
	client := b2.NewClient(nil)
	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <bucket-name>")
}

func TestCreateBucketCommand_RequiresValidBucketType(t *testing.T) {
	client := b2.NewClient(nil)
	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"-type=foo", "my-bucket"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, `-type must be either "public" or "private"`)
}

func TestCreateBucketCommand_CanCreateBucket(t *testing.T) {
	body := `{
		"accountId" : "010203040506",
		"bucketId" : "4a48fe8875c6214145260818",
		"bucketInfo" : {},
		"bucketName" : "my-bucket",
		"bucketType" : "allPrivate",
		"lifecycleRules" : []
	}`
	client := b2.NewClient(&testutil.FakeHTTPClient{
		Response: testutil.HTTPResponse(http.StatusOK, body),
	})
	client.Token = "TEST"

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"my-bucket"})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, fmt.Sprintf("Bucket %q created with ID %q", "my-bucket", "4a48fe8875c6214145260818"))
}
