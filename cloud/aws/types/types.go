package types

// Credential is an AWS credential from the credentials file
type Credential struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_session_token"`
}

// Profile is a local profile containing the credential and configuration.
type Profile struct {
	Name   string
	Cloud  string
	Cred   Credential
	// config    Config
	Source string
	IsCurrent bool
}

// Profiles is a collection of Profile
type Profiles struct {
	Profiles []Profile
}
