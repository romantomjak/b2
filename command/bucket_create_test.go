package command

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q, but it didn't", got, want)
	}
}

func TestCreateBucketCommand_RequiresBucketName(t *testing.T) {
	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui}

	code := cmd.Run([]string{})
	assertEqual(t, code, 1)

	out := ui.ErrorWriter.String()
	assertContains(t, out, "This command takes one argument")
}
