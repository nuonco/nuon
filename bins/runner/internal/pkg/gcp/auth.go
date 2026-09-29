package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gcptypes "github.com/nuonco/nuon/pkg/types/gcp"
)

func BuildMetadataRequest(ctx context.Context) (*gcptypes.MetadataRequest, error) {
	project, err := GetProjectID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID: %w", err)
	}

	zone, err := GetZone(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	instanceName, err := GetInstanceName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance name: %w", err)
	}

	tokenJSON, err := GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal([]byte(tokenJSON), &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse access token response: %w", err)
	}

	url := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s", project, zone, instanceName)

	return &gcptypes.MetadataRequest{
		Method: http.MethodGet,
		URL:    url,
		Headers: map[string]string{
			"Authorization": "Bearer " + tokenResp.AccessToken,
		},
	}, nil
}
