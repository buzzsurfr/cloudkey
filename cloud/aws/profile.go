package aws

import (
	"errors"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Profile is a local profile containing the credential and configuration.
type Profile struct {
	Name  string
	Cloud string
	Cred  Credential
	// config    Config
	Source    string
	IsCurrent bool
}

// Profiles is a collection of Profile
type Profiles struct {
	Profiles []Profile
}

// Session creates an AWS session
func (p *Profile) Session() *session.Session {
	return session.Must(session.NewSession())
}

// RotateKey creates a new key and deletes the old key (using the new key)
func (p *Profile) RotateKey() (bool, error) {
	// TODO
	return false, nil
}

// FromEnviron gets a profile from the credential environment variables
func FromEnviron() (Profile, error) {
	if c, ok := getCredentialFromEnviron(); ok {
		return Profile{
			Name:      "",
			Cloud:     "aws",
			Cred:      c,
			Source:    "EnvironmentVariable",
			IsCurrent: true,
		}, nil
	}
	return Profile{}, errors.New("No credential found in environment variable")
}

// FromConfigFile gets a list of profiles from the configuration file (default path/file is ~/.aws/credentials)
func FromConfigFile(findDefault bool) (Profiles, error) {
	var profiles Profiles

	awsConfigPath, err := getConfigPath()
	if err != nil {
		return profiles, err // Returning profiles since it's empty here
		// May change later since this assumes no credentials file found in ~/.aws
	}

	// Parse AWS config file
	v := viper.New()
	v.SetConfigName("credentials")
	v.SetConfigType("ini")
	v.AddConfigPath(awsConfigPath)
	err = v.ReadInConfig()
	if err != nil {
		return profiles, err // Returning profiles since it's empty here
	}
	// fmt.Printf("Configuration file:\n%+v\n\n", v) // DEBUG
	allSettings := v.AllSettings()

	var currentProfile string
	if findDefault {
		currentProfile = getCurrentProfile()
	}

	for key, value := range allSettings {
		// fmt.Printf("Key: %s\nValue: %+v\n\n", key, value) // DEBUG
		var cred Credential
		mapstructure.Decode(value, &cred)
		profiles.Profiles = append(profiles.Profiles, Profile{
			Name:      key,
			Cloud:     "aws",
			Cred:      cred,
			Source:    "ConfigFile",
			IsCurrent: findDefault && currentProfile == key,
		})
	}

	// Sort by profile name
	sort.Slice(profiles.Profiles, func(i, j int) bool { return profiles.Profiles[i].Name < profiles.Profiles[j].Name })

	return profiles, err
}

func getConfigPath() (string, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	hde, _ := homedir.Expand(hd)
	if err != nil {
		return "", err
	}
	configPath := hde + string(os.PathSeparator) + ".aws"
	// fmt.Printf("AWS Config File directory: %s\n", configPath) // DEBUG
	return configPath, nil
}

func getCurrentProfile() string {
	currentProfile := "default"

	// Check environment variables
	profileVars := []string{
		"AWS_DEFAULT_PROFILE",
		"AWS_PROFILE",
	}
	for _, env := range profileVars {
		v, ok := os.LookupEnv(env)
		if ok {
			currentProfile = v
		}

	}
	return currentProfile
}
