package aws

import (
	"os"
	"reflect"
	"testing"
)

func TestGetCredentialFromEnviron(t *testing.T) {
	accessKeyID := "AKIAIOSFODNN7EXAMPLE"
	secretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

	t.Run("valid credentials", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)

		got, ok := getCredentialFromEnviron()
		want := Credential{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}

		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
	t.Run("temporary session returns empty", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", "testing")

		got, ok := getCredentialFromEnviron()
		want := Credential{}

		if ok || !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
	t.Run("no environment variables", func(t *testing.T) {
		got, ok := getCredentialFromEnviron()
		want := Credential{}

		if ok || !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}

func TestFromAccessKey(t *testing.T) {
	return
}
