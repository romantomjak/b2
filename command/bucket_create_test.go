package command

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	b2 "github.com/romantomjak/b2/api"
)

var (
	mux    *http.ServeMux
	client *b2.Client
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = b2.NewClient(&b2.ApplicationCredentials{"1234", "MYSECRET"})
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q, but it didn't", got, want)
	}
}

func assertHttpMethod(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestCreateBucketCommand_RequiresBucketName(t *testing.T) {
	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	assertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	assertContains(t, out, "This command takes one argument: <bucket-name>")
}

func TestCreateBucketCommand_RequiresValidBucketType(t *testing.T) {
	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"-type=foo", "my-bucket"})
	assertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	assertContains(t, out, `-type must be either "public" or "private"`)
}

func TestCreateBucketCommand_CanCreateBucket(t *testing.T) {
	setup()
	defer teardown()

	// manually set the token so the client does not perform authentication
	// FIXME: maybe extract into a performTestAuthentication()?
	client.Token = "TEST"

	mux.HandleFunc("/b2api/v2/b2_create_bucket", func(w http.ResponseWriter, r *http.Request) {
		assertHttpMethod(t, r.Method, "POST")

		fmt.Fprint(w, `{
			"accountId" : "010203040506",
			"bucketId" : "4a48fe8875c6214145260818",
			"bucketInfo" : {},
			"bucketName" : "my-bucket",
			"bucketType" : "allPrivate",
			"lifecycleRules" : []
		}`)
	})

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"my-bucket"})
	assertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	assertContains(t, out, fmt.Sprintf("Bucket %q created with ID %q", "my-bucket", "4a48fe8875c6214145260818"))
}
