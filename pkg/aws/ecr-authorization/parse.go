package ecr

import (
	"encoding/base64"
	"fmt"
	"strings"

	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// parseAuthorizationData: parses authorization data into the required return format
func ParseAuthorizationData(data *ecr_types.AuthorizationData) (*Authorization, error) {
	auth, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("unable to decode auth string: %w", err)
	}

	authPieces := strings.SplitN(string(auth), ":", 2)
	return &Authorization{
		RegistryToken: authPieces[1],
		Username:      authPieces[0],
		ServerAddress: *data.ProxyEndpoint,
	}, nil
}
