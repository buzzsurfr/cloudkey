package aws

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	accessKeyID     = "AKIAIOSFODNN7EXAMPLE"
	secretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
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
	t.Run("valid access key", func(t *testing.T) {
		got, err := FromAccessKey(iam.AccessKey{
			AccessKeyId:     aws.String(accessKeyID),
			CreateDate:      aws.Time(time.Now()),
			SecretAccessKey: aws.String(secretAccessKey),
			Status:          aws.String("Active"),
			UserName:        aws.String("ValidUserName"),
		})
		want := Credential{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}

		if err != nil {
			t.Errorf("error: %v", err)
		} else if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}
