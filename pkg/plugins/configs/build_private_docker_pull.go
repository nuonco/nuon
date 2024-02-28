package configs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type encodedAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func PrivateDockerPullBuildEncodedAuthString(username, password string) (string, error) {
	auth := &encodedAuth{
		Username: username,
		Password: password,
	}
	byts, err := json.Marshal(auth)
	if err != nil {
		return "", fmt.Errorf("unable to marshal json auth: %w", err)
	}

	encodedByts := base64.StdEncoding.EncodeToString(byts)
	return encodedByts, nil
}

// PrivateDockerPullBuild is used to reference a private docker build
type PrivateDockerPullBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image string `hcl:"image"`
	Tag   string `hcl:"tag"`

	// the encoded auth that we use throughout our code looks something like the following,
	EncodedAuth       string `hcl:"encoded_auth"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint,optional"`
}
