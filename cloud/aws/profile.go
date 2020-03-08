package aws

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"gopkg.in/ini.v1"

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
	sts.GetCallerIdentityOutput
	Session *session.Session
	STS     stsiface.STSAPI
}

// Profiles is a collection of Profile
type Profiles struct {
	Profiles []Profile
}

// String formats the profile's attributes to a string
func (p *Profile) String() string {
	return fmt.Sprintf("Name: %s\nCloud: %s\nAccess Key: %s\nSource: %s\nAccount: %s\nArn: %s\n", p.Name, p.Cloud, p.Cred.AccessKeyID, p.Source, aws.StringValue(p.Account), aws.StringValue(p.Arn))
}

// NewSession creates an AWS session
func (p *Profile) NewSession() error {
	switch p.Source {
	case "EnvironmentVariable":
		// fmt.Println("Source: EnvironmentVariable")
		p.Session = session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewEnvCredentials(),
		}))
		return nil
	case "ConfigFile":
		// fmt.Println("Source: ConfigFile")
		p.Session = session.Must(session.NewSessionWithOptions(session.Options{
			Profile: p.Name,
		}))
		return nil
	}
	return ErrUnknownSource
}

// NewSTS creates a new STS client from the current session
func (p *Profile) NewSTS() {
	p.STS = sts.New(p.Session)
}

// RotateKey creates a new key and deletes the old key (using the new key)
// func (p *Profile) RotateKey() (bool, error) {
// 	// TODO
// 	return false, nil
// }

// Current gets the current profile
func Current() (Profile, error) {
	envProfile, err := FromEnviron()
	if err == nil { // we found a profile in env
		return envProfile, err
	}
	// Didn't find profile in environment variable, get profile from config file
	configProfiles, err := FromConfigFile(true)
	if err == nil { // we found profile(s) in config file
		for _, p := range configProfiles.Profiles {
			if p.IsCurrent {
				return p, err
			}
		}
	}
	// Didn't find profile in either environment variable or config file, return error
	return Profile{}, ErrCredentialNotFound
}

// GetByName gets the profile by name
func GetByName(profileName string) (Profile, error) {
	configProfiles, err := FromConfigFile(true)
	if err == nil { // we found profile(s) in config file
		for _, p := range configProfiles.Profiles {
			if p.Name == profileName {
				return p, err
			}
		}
	}
	// Didn't find profile in either environment variable or config file, return error
	return Profile{}, errors.New("No credential with profile name " + profileName + " found")
}

// FromEnviron gets a profile from the credential environment variables
func FromEnviron() (Profile, error) {
	if c, ok := getCredentialFromEnviron(); ok {
		return Profile{
			Name:                    "",
			Cloud:                   "aws",
			Cred:                    c,
			Source:                  "EnvironmentVariable",
			IsCurrent:               true,
			GetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
		}, nil
	}
	return Profile{}, ErrCredentialNotFound
}

// FromConfigFile gets a list of profiles from the configuration file (default path/file is ~/.aws/credentials)
func FromConfigFile(findDefault bool) (Profiles, error) {
	var profiles Profiles

	awsConfigPath, err := getConfigPath()
	if err != nil {
		return profiles, err // Returning profiles since it's empty here
		// May change later since this assumes no credentials file found in ~/.aws
	}

	// Determine whether to highlight a current profile
	var currentProfile string
	switch findDefault {
	case true:
		currentProfile = getCurrentProfile()
	case false:
		currentProfile = ""
	}

	// Parse AWS config file
	profiles, err = parseConfigFile(filepath.Join(awsConfigPath, "credentials"), currentProfile)
	if err != nil {
		return Profiles{}, err
	}

	return profiles, err
}

func parseConfigFile(path string, defaultProfile string) (Profiles, error) {
	var profiles Profiles

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("ini")
	err := v.ReadInConfig()
	if err != nil {
		return profiles, err // Returning profiles since it's empty here
	}

	// fmt.Printf("Configuration file:\n%+v\n\n", v) // DEBUG
	allSettings := v.AllSettings()
	// fmt.Printf("AllSettings: %+v\n", allSettings)

	for key, value := range allSettings {
		// fmt.Printf("Key: %s\nValue: %+v\n\n", key, value) // DEBUG
		var cred Credential
		mapstructure.Decode(value, &cred)
		profiles.Profiles = append(profiles.Profiles, Profile{
			Name:      key,
			Cloud:     "aws",
			Cred:      cred,
			Source:    "ConfigFile",
			IsCurrent: defaultProfile == key,
		})
	}

	return profiles, nil
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

// Lookup adds metadata from the cloud about the current proile
func (p *Profile) Lookup() error {
	// AWS sts:GetCallerIdentity API
	result, err := p.STS.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return err
		}
		return err
	}

	// Parse ARN
	resultArn, err := arn.Parse(*result.Arn)
	if err != nil {
		return err
	}

	// Verify is a user
	s := strings.Split(resultArn.Resource, "/")
	if s[0] != "user" {
		return ErrUnsupportedIdentityType
	}

	p.GetCallerIdentityOutput = *result

	return nil
}

// UpdateCredential locally updates the credential based on the profile type
func (p *Profile) UpdateCredential(cred Credential) error {
	switch p.Source {
	case "EnvironmentVariable":
		os.Setenv("AWS_ACCESS_KEY_ID", cred.AccessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", cred.SecretAccessKey)
	case "ConfigFile":
		akidErr := exec.Command("aws", "--profile", p.Name, "configure", "set", "aws_access_key_id", cred.AccessKeyID).Run()
		if akidErr != nil {
			panic(akidErr)
		}
		asakErr := exec.Command("aws", "--profile", p.Name, "configure", "set", "aws_secret_access_key", cred.SecretAccessKey).Run()
		if asakErr != nil {
			panic(asakErr)
		}
	}
	p.Cred = cred

	return nil
}

// WriteConfig writes the profiles to the default config file
func (p *Profiles) WriteConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	fileName := "credentials"
	return p.WriteConfigAs(configPath + string(os.PathSeparator) + fileName)
}

// WriteConfigAs writes the profiles to the config filename specified
func (p *Profiles) WriteConfigAs(filename string) error {
	// Marshal to INI file type
	cfg := ini.Empty()
	ini.PrettyFormat = false
	ini.PrettyEqual = true
	for _, profile := range p.Profiles {
		if profile.Source != "ConfigFile" {
			continue
		}
		cfg.Section(profile.Name).Key("aws_access_key_id").SetValue(profile.Cred.AccessKeyID)
		cfg.Section(profile.Name).Key("aws_secret_access_key").SetValue(profile.Cred.SecretAccessKey)
		if profile.Cred.SessionToken != "" {
			cfg.Section(profile.Name).Key("aws_session_token").SetValue(profile.Cred.SessionToken)
		}
	}

	// Save to file
	err := cfg.SaveTo(filename)
	if err != nil {
		return err
	}
	return nil
}
