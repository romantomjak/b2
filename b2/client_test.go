package b2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/romantomjak/b2/testutil"
)

func TestClient_Authorization(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/b2api/v2/b2_authorize_account", func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertHttpMethod(t, r.Method, "GET")

		fmt.Fprint(w, `{
			"absoluteMinimumPartSize": 5000000,
			"accountId": "abc123",
			"allowed": {
			  "bucketId": "my-bucket",
			  "bucketName": "MY BUCKET",
			  "capabilities": ["listBuckets","listFiles","readFiles","shareFiles","writeFiles","deleteFiles"],
			  "namePrefix": null
			},
			"apiUrl": "https://api123.backblazeb2.com",
			"authorizationToken": "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=",
			"downloadUrl": "https://f123.backblazeb2.com",
			"recommendedPartSize": 100000000
		}`)
	})
	server := httptest.NewServer(mux)
	client, err := NewClient(SetBaseURL(server.URL))

	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, client.AccountID, "abc123")
}

func TestClient_NewRequestDefauls(t *testing.T) {
	client, _ := NewClient(SetAuthentication(&Authorization{
		AbsoluteMinimumPartSize: 5000000,
		AccountID:               "abc123",
		Allowed: TokenCapability{
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

	inBody := map[string]string{"foo": "bar", "hello": "world"}
	outBody := `{"foo":"bar","hello":"world"}` + "\n"
	req, _ := client.NewRequest(http.MethodPost, "foo", inBody)

	// test relative URL was expanded
	testutil.AssertEqual(t, req.URL.String(), "https://api123.backblazeb2.com/foo")

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	testutil.AssertEqual(t, userAgent[:2], "b2")

	// test authorization token is attached to the request
	authToken := req.Header.Get("Authorization")
	testutil.AssertEqual(t, authToken, "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=")

	// test body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	testutil.AssertEqual(t, string(body), outBody)
}

func TestClient_APIErrorsAreReportedToUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/b2api/v2/b2_authorize_account", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)

		fmt.Fprint(w, `{
			"status" : 401,
			"code" : "unauthorized",
			"message" : "The applicationKeyId and/or the applicationKey are wrong."
		}`)
	})
	server := httptest.NewServer(mux)
	_, err := NewClient(SetBaseURL(server.URL))

	testutil.AssertNotNil(t, err)
	testutil.AssertContains(t, err.Error(), "The applicationKeyId and/or the applicationKey are wrong.")
}
