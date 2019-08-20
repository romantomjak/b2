package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestCreateBucketCommand_RequiresBucketName(t *testing.T) {
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
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <bucket-name>")
}

func TestCreateBucketCommand_RequiresValidBucketType(t *testing.T) {
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
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"-type=foo", "my-bucket"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, `-type must be either "public" or "private"`)
}

func TestCreateBucketCommand_BucketCreateRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/b2api/v2/b2_create_bucket", func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertHttpMethod(t, r.Method, "POST")

		var got map[string]string
		json.NewDecoder(r.Body).Decode(&got)

		testutil.AssertEqual(t, got["accountId"], "abc123")
		testutil.AssertEqual(t, got["bucketName"], "my-bucket")
		testutil.AssertEqual(t, got["bucketType"], "allPrivate")
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
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	cmd.Run([]string{"my-bucket"})
}

func TestCreateBucketCommand_PrintsCreatedBucketID(t *testing.T) {
	mux := http.NewServeMux()
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
	cmd := &CreateBucketCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"my-bucket"})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, fmt.Sprintf("Bucket %q created with ID %q", "my-bucket", "4a48fe8875c6214145260818"))
}
