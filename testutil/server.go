package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// NewServer starts and returns a new HTTP Server.
//
// It is used for end-to-end HTTP tests and is preconfigured to return
// a B2 API authorization response. Users should attach their own HandleFunc
// functions to the returned mux.
//
// It is callers responsibility to call Close when finished, to shut it down
func NewServer() (*httptest.Server, *http.ServeMux) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	authJSON := `{
		"absoluteMinimumPartSize": 5000000,
		"accountId": "abc123",
		"allowed": {
		  "bucketId": "my-bucket",
		  "bucketName": "MY BUCKET",
		  "capabilities": ["listBuckets","listFiles","readFiles","shareFiles","writeFiles","deleteFiles"],
		  "namePrefix": null
		},
		"apiUrl": "%s",
		"authorizationToken": "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=",
		"downloadUrl": "%s",
		"recommendedPartSize": 100000000
	}`
	mux.HandleFunc("/b2api/v2/b2_authorize_account", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, authJSON, server.URL, server.URL)
	})
	return server, mux
}
