package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	mux    *http.ServeMux
	client *Client
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(&ApplicationCredentials{"1234", "MYSECRET"})
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func assertStrings(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertHttpMethod(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestClient_NewClientDefaultValues(t *testing.T) {
	c := NewClient(&ApplicationCredentials{"1234", "MYSECRET"})
	assertStrings(t, c.UserAgent[:2], "b2")
	assertStrings(t, c.BaseURL.String(), "https://api.backblazeb2.com/")
}

func TestClient_NewRequestHeaders(t *testing.T) {
	c := NewClient(&ApplicationCredentials{"1234", "MYSECRET"})
	c.Token = "TEST"

	req, _ := c.NewRequest(http.MethodGet, "foo")

	// test relative URL was expanded
	assertStrings(t, req.URL.String(), "https://api.backblazeb2.com/foo")

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	assertStrings(t, c.UserAgent, userAgent)

	// test authorization token is attached to the request
	authToken := req.Header.Get("Authorization")
	assertStrings(t, authToken, "TEST")
}

func TestClient_AcquiresNewTokenWhenTokenIsNotSet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+authorizeAccountURL, func(w http.ResponseWriter, r *http.Request) {
		assertHttpMethod(t, r.Method, "POST")

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
			"downloadUrl": "https://f002.backblazeb2.com",
			"recommendedPartSize": 100000000
		}`)
	})

	req, _ := client.NewRequest(http.MethodGet, "foo")

	authToken := req.Header.Get("Authorization")
	assertStrings(t, authToken, "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=")
}
