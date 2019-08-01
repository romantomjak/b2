package config

import (
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestConfig_FromEnv(t *testing.T) {
	c := FromEnv([]string{"B2_KEY_ID=mykey", "B2_KEY_SECRET=muchsecret", "FOO=BAR=1"})

	assertEqual(t, c.ApplicationKeyID, "mykey")
	assertEqual(t, c.ApplicationKeySecret, "muchsecret")
}
