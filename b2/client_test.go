package b2

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romantomjak/b2/testutil"
)

func TestClient_Authorization(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	cache, err := NewInMemoryCache()
	require.NoError(t, err)

	client, err := NewClient("key-id", "key-secret", SetBaseURL(server.URL), SetCache(cache))
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClient_AuthorizationCache(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	tmpDir, err := ioutil.TempDir(os.TempDir(), "b2-cli-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewDiskCache(tmpDir)
	require.NoError(t, err)

	timeNow = func() time.Time {
		return time.Date(2020, 10, 21, 22, 48, 0, 0, time.UTC)
	}

	authJSON := `{
		"authorization": {
			"tokenExpiresAt": "2020-10-22T22:48:00Z",
			"absoluteMinimumPartSize": 5000000,
			"accountId": "abc123",
			"allowed": {
			"bucketId": "my-bucket",
			"bucketName": "MY BUCKET",
			"capabilities": ["listBuckets","listFiles","readFiles","shareFiles","writeFiles","deleteFiles"],
			"namePrefix": ""
			},
			"apiUrl": "%s",
			"authorizationToken": "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=",
			"downloadUrl": "%s",
			"recommendedPartSize": 100000000
		}
	}`

	_, err = NewClient("key-id", "key-secret", SetBaseURL(server.URL), SetCache(cache))
	require.NoError(t, err)

	cacheFile := filepath.Join(tmpDir, "cache")
	authBytes, err := ioutil.ReadFile(cacheFile)
	assert.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(authJSON, server.URL, server.URL), string(authBytes))
}

func TestClient_NewRequestDefaults(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	cache, err := NewInMemoryCache()
	require.NoError(t, err)

	client, err := NewClient("key-id", "key-secret", SetBaseURL(server.URL), SetCache(cache))
	require.NoError(t, err)

	inBody := map[string]string{"foo": "bar", "hello": "world"}
	outBody := `{"foo":"bar","hello":"world"}` + "\n"
	ctx := context.TODO()
	req, err := client.NewRequest(ctx, http.MethodPost, "foo", inBody)
	require.NoError(t, err)

	// test relative URL was expanded
	absURL := fmt.Sprintf("%s/%s", server.URL, "foo")
	assert.Equal(t, absURL, req.URL.String())

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	assert.Equal(t, "b2", userAgent[:2])

	// test authorization token is attached to the request
	authToken := req.Header.Get("Authorization")
	assert.Equal(t, "4_0022623512fc8f80000000001_0186e431_d18d02_acct_tH7VW03boebOXayIc43-sxptpfA=", authToken)

	// test body was JSON encoded
	body, err := ioutil.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, outBody, string(body))
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
	defer server.Close()

	cache, err := NewInMemoryCache()
	require.NoError(t, err)

	_, err = NewClient("key-id", "key-secret", SetBaseURL(server.URL), SetCache(cache))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The applicationKeyId and/or the applicationKey are wrong.")
}
