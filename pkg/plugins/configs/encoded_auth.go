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
