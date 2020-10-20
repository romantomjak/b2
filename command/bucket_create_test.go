package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCreateBucketCommand_RequiresBucketName(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	code := cmd.Run([]string{})
	assert.Equal(t, 1, code)

	out := ui.ErrorWriter.String()
	assert.Contains(t, out, "This command takes one argument: <bucket-name>")
}

func TestCreateBucketCommand_RequiresValidBucketType(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	code := cmd.Run([]string{"-type=foo", "my-bucket"})
	assert.Equal(t, 1, code)

	out := ui.ErrorWriter.String()
	assert.Contains(t, out, `-type must be either "public" or "private"`)
}

func TestCreateBucketCommand_BucketCreateRequest(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	mux.HandleFunc("/b2api/v2/b2_create_bucket", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var got map[string]string
		json.NewDecoder(r.Body).Decode(&got)

		assert.Equal(t, "abc123", got["accountId"])
		assert.Equal(t, "my-bucket", got["bucketName"])
		assert.Equal(t, "allPrivate", got["bucketType"])
	})

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	_ = cmd.Run([]string{"my-bucket"})
	// return code is ignored on purpose here.
	// fake b2_create_bucket handler is not writing the response, so
	// the command will fail and return 1
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

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	code := cmd.Run([]string{"my-bucket"})
	assert.Equal(t, 0, code)

	out := ui.OutputWriter.String()
	assert.Contains(t, out, fmt.Sprintf("Bucket %q created with ID %q", "my-bucket", "4a48fe8875c6214145260818"))
}
