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

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &GetCommand{Ui: ui, Client: client}

	src := fmt.Sprintf("my-bucket/%s", filepath.Base(tmpFile.Name()))
	dst := "/tmp/testing.txt"
	defer os.Remove(dst)

	code := cmd.Run([]string{src, dst})
	testutil.AssertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	testutil.AssertContains(t, out, fmt.Sprintf("Downloaded %s to %s", src, dst))
}
