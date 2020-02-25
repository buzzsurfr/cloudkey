package aws

import "os"

// Credential is an AWS credential from the credentials file
type Credential struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_session_token"`
}

func getCredentialFromEnviron() (Credential, bool) {
	if _, snok := os.LookupEnv("AWS_SESSION_NAME"); snok {
		return Credential{}, false
	}
	akid, akok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	sak, skok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if akok && skok {
		return Credential{
			AccessKeyID:     akid,
			SecretAccessKey: sak,
		}, true
	}

	return Credential{}, false
}
