package command

import (
	"fmt"
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
	assertContains(t, out, "This command takes one argument: <bucket-name>")
}

func TestCreateBucketCommand_RequiresValidBucketType(t *testing.T) {
	testCases := []struct {
		bucketType string
		exitCode   int
		stdErr     string
	}{
		{"foo", 1, `-type must be either "public" or "private"`},
		{"public", 0, ""},
		{"private", 0, ""},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("type=%s", tc.bucketType), func(t *testing.T) {
			ui := cli.NewMockUi()
			cmd := &CreateBucketCommand{Ui: ui}

			code := cmd.Run([]string{"-type=" + tc.bucketType, "my-bucket"})
			assertEqual(t, code, tc.exitCode)

			out := ui.ErrorWriter.String()
			assertContains(t, out, tc.stdErr)
		})
	}
}

func TestCreateBucketCommand_PrintsSuccessMessage(t *testing.T) {
	ui := cli.NewMockUi()
	cmd := &CreateBucketCommand{Ui: ui}

	code := cmd.Run([]string{"my-bucket"})
	assertEqual(t, code, 0)

	out := ui.OutputWriter.String()
	assertContains(t, out, fmt.Sprintf("Successfully created %q Bucket!", "my-bucket"))
}
