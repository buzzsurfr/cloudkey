package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func TestParseConfigFile(t *testing.T) {
	profileName := "default"
	tempConfig := fmt.Sprintf("[%s]\naws_access_key_id = %s\naws_secret_access_key = %s\n", profileName, accessKeyID, secretAccessKey)
	// Mock config file
	tempConfigFile, err := ioutil.TempFile("", "awsconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempConfigFile.Name())

	if _, err := tempConfigFile.Write([]byte(tempConfig)); err != nil {
		t.Fatal(err)
	}

	if _, err := tempConfigFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	t.Run("working config file", func(t *testing.T) {

		got, err := parseConfigFile(tempConfigFile.Name(), false)
		want := Profiles{
			Profiles: []Profile{
				Profile{
					Name:  profileName,
					Cloud: "aws",
					Cred: Credential{
						AccessKeyID:     accessKeyID,
						SecretAccessKey: secretAccessKey,
					},
					Source:    "ConfigFile",
					IsCurrent: false,
				},
			},
		}

		assertProfiles(t, got, want)
		assertNoError(t, err)
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("default path", func(t *testing.T) {

		got, gotErr := getConfigPath()
		want, wantErr := homedir.Expand("~/.aws")

		assertString(t, got, want)
		assertNoError(t, gotErr)
		assertNoError(t, wantErr)
	})
}

func TestGetCurrentProfile(t *testing.T) {
	defaultProfile := "otherdefault"
	testProfile := "test"
	t.Run("no profile set", func(t *testing.T) {
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Unsetenv("AWS_PROFILE")

		got := getCurrentProfile()
		want := "default"

		assertString(t, got, want)
	})
	t.Run("AWS_DEFAULT_PROFILE set", func(t *testing.T) {
		os.Setenv("AWS_DEFAULT_PROFILE", defaultProfile)
		os.Unsetenv("AWS_PROFILE")

		got := getCurrentProfile()
		want := defaultProfile

		assertString(t, got, want)
	})
	t.Run("AWS_PROFILE set", func(t *testing.T) {
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Setenv("AWS_PROFILE", testProfile)

		got := getCurrentProfile()
		want := testProfile

		assertString(t, got, want)
	})
	t.Run("AWS_PROFILE overrides AWS_DEFAULT_PROFILE", func(t *testing.T) {
		os.Setenv("AWS_DEFAULT_PROFILE", defaultProfile)
		os.Setenv("AWS_PROFILE", testProfile)

		got := getCurrentProfile()
		want := testProfile

		assertString(t, got, want)
	})
}

func assertString(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q but want %q", got, want)
	}
}

func assertProfiles(t *testing.T, got, want Profiles) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func assertProfile(t *testing.T, got, want Profile) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func assertError(t *testing.T, got error, want error) {
	t.Helper()
	if got == nil {
		t.Fatal("wanted an error but didn't get one")
	}
	if got.Error() != want.Error() {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got error %q but didn't want one", err)
	}
}
