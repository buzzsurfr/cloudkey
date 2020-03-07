package cmd

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
)

// UserName gets the user name from the ARN (passed as string)
func UserName(a string) (string, error) {
	// Parse ARN
	resultArn, err := arn.Parse(a)
	if err != nil {
		return "", err
	}

	// Verify is a user
	s := strings.Split(resultArn.Resource, "/")
	if s[0] != "user" {
		return "", errors.New("Not a user")
	}
	userName := s[1]

	return userName, nil
}
