package ecr

import (
	"encoding/base64"
	"fmt"
	"strings"

	ecrpublictypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
)

func ParsePublicAuthorizationData(data *ecrpublictypes.AuthorizationData) (*Authorization, error) {
	auth, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("unable to decode auth string: %w", err)
	}

	authPieces := strings.SplitN(string(auth), ":", 2)
	return &Authorization{
		RegistryToken: authPieces[1],
		Username:      authPieces[0],
	}, nil
}
