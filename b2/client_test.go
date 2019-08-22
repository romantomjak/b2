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
	server, _ := testutil.NewServer()
	defer server.Close()

	client, err := NewClient(SetBaseURL(server.URL))

	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, client.AccountID, "abc123")
}

func TestClient_NewRequestDefauls(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	client, _ := NewClient(SetBaseURL(server.URL))

	inBody := map[string]string{"foo": "bar", "hello": "world"}
	outBody := `{"foo":"bar","hello":"world"}` + "\n"
	req, _ := client.NewRequest(http.MethodPost, "foo", inBody)

	// test relative URL was expanded
	absURL := fmt.Sprintf("%s/%s", server.URL, "foo")
	testutil.AssertEqual(t, req.URL.String(), absURL)

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
