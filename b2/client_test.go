package b2

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/romantomjak/b2/testutil"
)

func TestClient_NewClientDefaults(t *testing.T) {
	client := NewClient(nil)
	testutil.AssertEqual(t, client.UserAgent[:2], "b2")
	testutil.AssertEqual(t, client.BaseURL.String(), "https://api.backblazeb2.com/")
}

func TestClient_NewRequestDefauls(t *testing.T) {
	client := NewClient(nil)
	client.Token = "TEST"

	inBody := map[string]string{"foo": "bar", "hello": "world"}
	outBody := `{"foo":"bar","hello":"world"}` + "\n"
	req, _ := client.NewRequest(http.MethodPost, "foo", inBody)

	// test relative URL was expanded
	testutil.AssertEqual(t, req.URL.String(), "https://api.backblazeb2.com/foo")

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	testutil.AssertEqual(t, client.UserAgent, userAgent)

	// test authorization token is attached to the request
	authToken := req.Header.Get("Authorization")
	testutil.AssertEqual(t, authToken, "TEST")

	// test body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	testutil.AssertEqual(t, string(body), outBody)
}

func TestClient_NewRequestAuthentication(t *testing.T) {
	body := `{
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
		"downloadUrl": "https://f002.backblazeb2.com",
		"recommendedPartSize": 100000000
	}`
	client := NewClient(&testutil.FakeHTTPClient{
		Response: testutil.HTTPResponse(http.StatusOK, body),
	})

	// the HTTP method here is irrevelant because the authentication call will
	// be issued before the prepared request is returned
	req, _ := client.NewRequest(http.MethodGet, "foo", nil)

	// test authorization token is set
	authToken := req.Header.Get("Authorization")
	testutil.AssertEqual(t, authToken, "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=")

	// test base url from the authorization response is set
	testutil.AssertEqual(t, client.BaseURL.String(), "https://api123.backblazeb2.com")

	// test account id is set
	testutil.AssertEqual(t, client.AccountID, "abc123")
}

func TestClient_APIErrorsAreReportedToUser(t *testing.T) {
	body := `{
		"status" : 401,
		"code" : "unauthorized",
		"message" : "The applicationKeyId and/or the applicationKey are wrong."
	}`
	client := NewClient(&testutil.FakeHTTPClient{
		Response: testutil.HTTPResponse(http.StatusUnauthorized, body),
	})

	// the HTTP method here is irrevelant because the authentication call will
	// be issued before the prepared request is returned
	_, err := client.NewRequest(http.MethodGet, "foo", nil)

	// test authorization error is reported to the user
	testutil.AssertContains(t, err.Error(), "The applicationKeyId and/or the applicationKey are wrong.")
}
