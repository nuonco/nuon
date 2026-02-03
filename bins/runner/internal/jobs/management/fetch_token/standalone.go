package fetchtoken

import (
	"context"
	"fmt"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	pkgaws "github.com/nuonco/nuon/bins/runner/internal/pkg/aws"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
)

type FetchTokenResult struct {
	RunnerID   string
	InstanceID string
	AccountID  string
	TokenPath  string
}

func FetchAndStoreToken(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
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

	if err := monitor.WriteRunnerTokenFile(resp.Token); err != nil {
		return nil, fmt.Errorf("failed to write token: %w", err)
	}

	return &FetchTokenResult{
		RunnerID:   resp.RunnerID,
		InstanceID: resp.InstanceID,
		AccountID:  resp.AccountID,
		TokenPath:  monitor.RunnerTokenFilename,
	}, nil
}
