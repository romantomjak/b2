package client

import (
	"net/http"
	"testing"

	"github.com/romantomjak/b2/config"
)

func assertStrings(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestClient_NewClient(t *testing.T) {
	cfg := config.FromEnv([]string{"B2_KEY_ID=mykey", "B2_KEY_SECRET=muchsecret"})
	c := NewClient(nil)
	assertStrings(t, c.UserAgent[:2], "b2")
	assertStrings(t, c.BaseURL.String(), cfg.AuthorizationBaseURL.String())
}

func TestClient_NewRequest(t *testing.T) {
	cfg := config.FromEnv([]string{"B2_KEY_ID=mykey", "B2_KEY_SECRET=muchsecret"})
	c := NewClient(nil)

	req, _ := c.NewRequest(http.MethodGet, "foo", nil)

	// test relative URL was expanded
	assertStrings(t, req.URL.String(), cfg.AuthorizationBaseURL.String()+"foo")

	// // test body was JSON encoded
	// body, _ := ioutil.ReadAll(req.Body)
	// if string(body) != outBody {
	// 	t.Errorf("NewRequest(%v)Body = %v, expected %v", inBody, string(body), outBody)
	// }

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	assertStrings(t, c.UserAgent, userAgent)
}
