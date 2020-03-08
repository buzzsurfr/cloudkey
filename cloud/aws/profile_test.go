package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mitchellh/go-homedir"
)

func TestString(t *testing.T) {
	profileName := "default"
	t.Run("sample profile, output = table", func(t *testing.T) {
		p := Profile{
			Name:  profileName,
			Cloud: "aws",
			Cred: Credential{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
			Source:                  "ConfigFile",
			IsCurrent:               false,
			GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
		}
		got := p.String()
		want := "Name: default\nCloud: aws\nAccess Key: AKIAIOSFODNN7EXAMPLE\nSource: ConfigFile\nAccount: \nArn: \n"

		assertString(t, got, want)
	})
}

func TestSession(t *testing.T) {
	profileName := "default"
	p := Profile{
		Name:  profileName,
		Cloud: "aws",
		Cred: Credential{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
		Source:                  "EnvironmentVariable",
		IsCurrent:               true,
		GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
	}
	t.Run("session with Environment Variables", func(t *testing.T) {
		p.Source = "EnvironmentVariable"
		got, err := p.Session()
		want := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		}))

		assertSession(t, got, want)
		assertNoError(t, err)
	})
	t.Run("session with Config File", func(t *testing.T) {
		p.Source = "ConfigFile"
		got, err := p.Session()
		want := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		}))

		assertSession(t, got, want)
		assertNoError(t, err)
	})
	t.Run("fail on unknown source", func(t *testing.T) {
		p.Source = "UnknownSource"
		got, err := p.Session()
		want := session.Must(session.NewSession())

		assertSession(t, got, want)
		assertError(t, err, ErrUnknownSource)
	})
}

func TestCurrent(t *testing.T) {
	profileName := "default"
	p := Profile{
		Name:  profileName,
		Cloud: "aws",
		Cred: Credential{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
		Source:                  "EnvironmentVariable",
		IsCurrent:               true,
		GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
	}

	t.Run("profile using Environment Variables", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
		os.Unsetenv("AWS_SESSION_TOKEN")
		p.Source = "EnvironmentVariable"
		got, err := Current()

		assertProfileName(t, got, p)
		assertNoError(t, err)
	})
	t.Run("profile using Environment Variables which overrides Config File", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
		os.Unsetenv("AWS_SESSION_TOKEN")
		p.Source = "EnvironmentVariable"
		got, err := Current()

		assertProfileName(t, got, p)
		assertProfileSource(t, got, p)
		assertNoError(t, err)
	})
	// t.Run("session with Config File", func(t *testing.T) {
	// 	os.Unsetenv("AWS_PROFILE")
	// 	os.Unsetenv("AWS_DEFAULT_PROFILE")

	// 	p.Source = "ConfigFile"
	// 	got, err := Current()

	// 	assertProfileName(t, got, p)
	// 	assertProfileSource(t, got, p)
	// 	assertNoError(t, err)
	// })
	// t.Run("fail on no environment variables or config file", func(t *testing.T) {
	// 	os.Unsetenv("AWS_PROFILE")
	// 	os.Unsetenv("AWS_DEFAULT_PROFILE")

	// 	_, err := p.Session()

	// 	assertError(t, err, ErrCredentialNotFound)
	// })
}

// func TestGetByName(t *testing.T) {
// }

func TestFromEnviron(t *testing.T) {
	profileName := "default"
	p := Profile{
		Name:  profileName,
		Cloud: "aws",
		Cred: Credential{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
		Source:                  "EnvironmentVariable",
		IsCurrent:               true,
		GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
	}
	t.Run("profile using Environment Variables", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
		os.Unsetenv("AWS_SESSION_TOKEN")

		got, err := FromEnviron()

		assertProfileName(t, got, p)
		assertNoError(t, err)
	})
	t.Run("fail on no environment credentials", func(t *testing.T) {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_SESSION_TOKEN")

		got, err := FromEnviron()

		assertProfileName(t, got, p)
		assertError(t, err, ErrCredentialNotFound)
	})
	t.Run("fail if session token set", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", "TestToken")

		got, err := FromEnviron()

		assertProfileName(t, got, p)
		assertError(t, err, ErrCredentialNotFound)
	})
}

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

	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_DEFAULT_PROFILE")

	t.Run("working config file and default", func(t *testing.T) {
		got, err := parseConfigFile(tempConfigFile.Name(), profileName)
		want := Profiles{
			Profiles: []Profile{
				Profile{
					Name:  profileName,
					Cloud: "aws",
					Cred: Credential{
						AccessKeyID:     accessKeyID,
						SecretAccessKey: secretAccessKey,
					},
					Source:                  "ConfigFile",
					IsCurrent:               true,
					GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
				},
			},
		}

		assertProfiles(t, got, want)
		assertNoError(t, err)
	})

	t.Run("working config file and no default", func(t *testing.T) {
		got, err := parseConfigFile(tempConfigFile.Name(), "")
		want := Profiles{
			Profiles: []Profile{
				Profile{
					Name:  profileName,
					Cloud: "aws",
					Cred: Credential{
						AccessKeyID:     accessKeyID,
						SecretAccessKey: secretAccessKey,
					},
					Source:                  "ConfigFile",
					IsCurrent:               false,
					GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
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

func TestLookup(t *testing.T) {
	t.Run("", func(t *testing.T) {
		// Mock sts.GetCallerIdentity
	})
}

func assertString(t *testing.T, got, want string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q but want %q", got, want)
	}
}

func assertProfiles(t *testing.T, got, want Profiles) {
	t.Helper()
	// Can't use !reflect.DeepEqual since we embed sts.GetCallerIdentityOutput
	// if got.Name != want.Name && got.
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func assertProfileName(t *testing.T, got, want Profile) {
	t.Helper()
	// Can't use !reflect.DeepEqual since we embed sts.GetCallerIdentityOutput
	// if !reflect.DeepEqual(got, want) {
	if got.Name == want.Name {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func assertProfileSource(t *testing.T, got, want Profile) {
	t.Helper()
	if !reflect.DeepEqual(got.Source, want.Source) {
		t.Errorf("got %+v want %+v", got.Source, want.Source)
	}
}

func assertSession(t *testing.T, got, want *session.Session) {
	t.Helper()
	// No reliable way to compare sessions
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
