package fetchtoken

import (
	"context"
	"fmt"
	"os"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgaws "github.com/nuonco/nuon/bins/runner/internal/pkg/aws"
	pkggcp "github.com/nuonco/nuon/bins/runner/internal/pkg/gcp"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/token"
)

type FetchTokenResult struct {
	RunnerID   string `json:"runner_id"`
	InstanceID string `json:"instance_id"`
	AccountID  string `json:"account_id,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
	Token      string `json:"token,omitempty"`
	TokenPath  string `json:"token_path,omitempty"`
}

// FetchToken authenticates using cloud instance credentials and returns the token without writing to disk.
// It detects the cloud provider from the CLOUD_PROVIDER env var, falling back to auto-detection.
func FetchToken(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
	provider := detectCloudProvider(ctx)
	switch provider {
	case "gcp":
		return fetchTokenGCP(ctx, apiClient)
	case "aws":
		return fetchTokenAWS(ctx, apiClient)
	default:
		return nil, fmt.Errorf("unsupported or undetected cloud provider: %s", provider)
	}
}

// FetchAndStoreToken authenticates using cloud instance credentials and writes the token to disk.
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

func detectCloudProvider(ctx context.Context) string {
	if p := os.Getenv("CLOUD_PROVIDER"); p != "" {
		return p
	}

	if pkggcp.IsGCPInstance(ctx) {
		return "gcp"
	}

	// default to AWS for backward compatibility
	return "aws"
}

func fetchTokenAWS(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
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

func fetchTokenGCP(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
	apiURL := os.Getenv("RUNNER_API_URL")
	if apiURL == "" {
		apiURL = "https://runner.nuon.co"
	}

	// Step 1: Get identity token (JWT) - proves who we are (project, SA, instance)
	identityToken, err := pkggcp.GetIdentityToken(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP identity token: %w", err)
	}

	// Step 2: Build metadata request - lets server independently read our instance
	// metadata (including nuon_runner_id). Mirrors the AWS presigned DescribeTags pattern.
	metadataReq, err := pkggcp.BuildMetadataRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build GCP metadata request: %w", err)
	}

	resp, err := apiClient.RunnerAuthGCP(ctx, &models.ServiceRunnerAuthGCPRequest{
		IdentityToken: &identityToken,
		Metadata: &models.GcpMetadataRequest{
			Method:  metadataReq.Method,
			URL:     metadataReq.URL,
			Headers: metadataReq.Headers,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with GCP: %w", err)
	}

	if !resp.Authenticated {
		return nil, fmt.Errorf("authentication failed: runner was not authenticated")
	}

	return &FetchTokenResult{
		RunnerID:   resp.RunnerID,
		InstanceID: resp.InstanceID,
		ProjectID:  resp.ProjectID,
		Token:      resp.Token,
	}, nil
}
