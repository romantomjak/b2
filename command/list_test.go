package command

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
)

func TestListCommand_AcceptsPathArgument(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	client, _ := b2.NewClient(b2.SetBaseURL(server.URL))
	ui := cli.NewMockUi()
	cmd := &ListCommand{Ui: ui, Client: client}

	code := cmd.Run([]string{"one", "two"})
	testutil.AssertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	testutil.AssertContains(t, out, "This command takes one argument: <path>")
}
