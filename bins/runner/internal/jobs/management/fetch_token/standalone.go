package fetchtoken

import (
	"context"
	"fmt"
	"os"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgaws "github.com/nuonco/nuon/bins/runner/internal/pkg/aws"
	pkgazure "github.com/nuonco/nuon/bins/runner/internal/pkg/azure"
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
// For AWS, authMethod selects the authentication strategy: "" / "iid" (default) for Instance
// Identity Document (requires runnerID), or "sts" for presigned STS requests.
func FetchToken(ctx context.Context, apiClient nuonrunner.Client, authMethod, runnerID string) (*FetchTokenResult, error) {
	provider := detectCloudProvider(ctx)
	switch provider {
	case "gcp":
		return fetchTokenGCP(ctx, apiClient)
	case "aws":
		switch authMethod {
		case "sts":
			return fetchTokenAWSSTS(ctx, apiClient)
		default:
			return fetchTokenAWSIID(ctx, apiClient, runnerID)
		}
	default:
		return nil, fmt.Errorf("unsupported or undetected cloud provider: %s", provider)
	}
}

// FetchTokenAzure authenticates with Azure managed identity and returns the token without writing to disk.
func FetchTokenAzure(ctx context.Context, apiClient nuonrunner.Client, runnerID string) (*FetchTokenResult, error) {
	f := &pkgazure.Fetcher{RunnerID: runnerID}
	result, err := f.FetchToken(ctx, apiClient)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with azure: %w", err)
	}

	return &FetchTokenResult{
		RunnerID:  result.RunnerID,
		AccountID: result.AccountID,
		Token:     result.Token,
	}, nil
}

// FetchAndStoreToken authenticates using cloud instance credentials and writes the token to disk.
func FetchAndStoreToken(ctx context.Context, apiClient nuonrunner.Client, authMethod, runnerID string) (*FetchTokenResult, error) {
	result, err := FetchToken(ctx, apiClient, authMethod, runnerID)
	if err != nil {
		return nil, err
	}

	return storeToken(result)
}

// FetchAndStoreTokenAzure authenticates with Azure managed identity and writes the token to disk.
func FetchAndStoreTokenAzure(ctx context.Context, apiClient nuonrunner.Client, runnerID string) (*FetchTokenResult, error) {
	result, err := FetchTokenAzure(ctx, apiClient, runnerID)
	if err != nil {
		return nil, err
	}

	return storeToken(result)
}

func storeToken(result *FetchTokenResult) (*FetchTokenResult, error) {
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

func fetchTokenAWSSTS(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
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

func fetchTokenAWSIID(ctx context.Context, apiClient nuonrunner.Client, runnerID string) (*FetchTokenResult, error) {
	if runnerID == "" {
		return nil, fmt.Errorf("runner ID is required for IID auth")
	}

	iid, err := pkgaws.GetInstanceIdentityDocument(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance identity document: %w", err)
	}

	req := pkgaws.BuildIIDAuthRequest(iid, runnerID)

	resp, err := apiClient.RunnerAuthAWSIID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with IID: %w", err)
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
