package aws

import "errors"

// ErrCredentialNotFound means no credential existed that can be rotated by cloudkey.
var ErrCredentialNotFound = errors.New("No rotatable credential found. You may need to run aws configure")

// ErrUnknownSource is for a source not configured in our AWS cloud provider
var ErrUnknownSource = errors.New("Unknown source in profile")