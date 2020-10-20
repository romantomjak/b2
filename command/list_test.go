package command

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/b2"
	"github.com/romantomjak/b2/testutil"
	"github.com/stretchr/testify/assert"
)

func TestListCommand_AcceptsPathArgument(t *testing.T) {
	server, _ := testutil.NewServer()
	defer server.Close()

	client, _ := b2.NewClient("key-id", "key-secret", b2.SetBaseURL(server.URL))

	ui := cli.NewMockUi()
	cmd := &ListCommand{
		baseCommand: &baseCommand{ui: ui, client: client},
	}

	code := cmd.Run([]string{"one", "two"})
	assert.Equal(t, 1, code)

	out := ui.ErrorWriter.String()
	assert.Contains(t, out, "This command takes one argument: <path>")
}
