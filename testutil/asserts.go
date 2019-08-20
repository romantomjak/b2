package testutil

import (
	"reflect"
	"strings"
	"testing"
)

func AssertNil(t *testing.T, val interface{}) {
	t.Helper()
	if val != nil {
		t.Fatalf("expected %+v to be nil, but it wasn't", val)
	}
}

func AssertNotNil(t *testing.T, val interface{}) {
	t.Helper()
	if val == nil {
		t.Fatalf("expected %+v to not be nil, but it was", val)
	}
}

func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func AssertNotEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func AssertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q, but it didn't", got, want)
	}
}

func AssertHttpMethod(t *testing.T, got, want string) {
	AssertEqual(t, got, want)
}
