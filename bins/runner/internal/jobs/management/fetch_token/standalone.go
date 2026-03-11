package fetchtoken

import (
	"context"
	"fmt"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	pkgaws "github.com/nuonco/nuon/bins/runner/internal/pkg/aws"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/token"
)

type FetchTokenResult struct {
	RunnerID   string `json:"runner_id"`
	InstanceID string `json:"instance_id"`
	AccountID  string `json:"account_id"`
	Token      string `json:"token,omitempty"`
	TokenPath  string `json:"token_path,omitempty"`
}

// FetchToken authenticates with AWS and returns the token without writing to disk.
func FetchToken(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
	stsRequest, err := pkgaws.GetPresignedSTSRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned STS request: %w", err)
	}

	tagsRequest, err := pkgaws.GetPresignedInstanceTagsRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned EC2 tags request: %w", err)
	}

	req := pkgaws.BuildAuthRequest(stsRequest, tagsRequest)

	resp, err := apiClient.RunnerAuthAWS(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with AWS: %w", err)
	}

	if !resp.Authenticated {
		return nil, fmt.Errorf("authentication failed: runner was not authenticated")
	}

	return &FetchTokenResult{
		RunnerID:   resp.RunnerID,
		InstanceID: resp.InstanceID,
		AccountID:  resp.AccountID,
		Token:      resp.Token,
	}, nil
}

// FetchAndStoreToken authenticates with AWS and writes the token to disk.
func FetchAndStoreToken(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
	result, err := FetchToken(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	if err := token.WriteFile(result.Token); err != nil {
		return nil, fmt.Errorf("failed to write token: %w", err)
	}

	result.TokenPath = token.Filename
	result.Token = "" // don't leak the token in the result when stored to disk
	return result, nil
}
