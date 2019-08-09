package client

import (
	"net/http"
	"testing"
)

func assertStrings(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestClient_NewClientDefaultValues(t *testing.T) {
	c := NewClient()
	assertStrings(t, c.UserAgent[:2], "b2")
	assertStrings(t, c.BaseURL.String(), "https://api.backblazeb2.com/")
}

func TestClient_NewRequestHeaders(t *testing.T) {
	c := NewClient()
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
