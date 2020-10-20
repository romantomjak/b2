package command

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand_CanDownloadFile(t *testing.T) {
	server, mux := testutil.NewServer()
	defer server.Close()

	tmpFile, _ := ioutil.TempFile(os.TempDir(), "b2-cli-test-")
	defer os.Remove(tmpFile.Name())

	tmpFile.Write([]byte("This file is not empty."))
	tmpFile.Close()

	path := fmt.Sprintf("/file/my-bucket/%s", filepath.Base(tmpFile.Name()))
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, tmpFile.Name())
	})

	cache, _ := b2.NewInMemoryCache()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL), b2.SetCache(cache))

	ui := cli.NewMockUi()
	cmd := &GetCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	// TODO: use ioutil.TempFile for destination
	src := fmt.Sprintf("my-bucket/%s", filepath.Base(tmpFile.Name()))
	dst := "/tmp/testing.txt"
	defer os.Remove(dst)

	code := cmd.Run([]string{src, dst})
	assert.Equal(t, 0, code)

	out := ui.OutputWriter.String()
	assert.Contains(t, out, fmt.Sprintf("Downloaded %s to %s", src, dst))
}
