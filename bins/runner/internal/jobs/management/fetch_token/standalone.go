package fetchtoken

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgaws "github.com/nuonco/nuon/bins/runner/internal/pkg/aws"
	pkgazure "github.com/nuonco/nuon/bins/runner/internal/pkg/azure"
	pkggcp "github.com/nuonco/nuon/bins/runner/internal/pkg/gcp"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/token"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/tokenfetcher"
)

type FetchTokenResult struct {
	RunnerID   string    `json:"runner_id"`
	InstanceID string    `json:"instance_id"`
	AccountID  string    `json:"account_id,omitempty"`
	ProjectID  string    `json:"project_id,omitempty"`
	Token      string    `json:"token,omitempty"`
	TokenPath  string    `json:"token_path,omitempty"`
	Audience   string    `json:"audience,omitempty"`
	ExpiresAt  time.Time `json:"-"`
}

const (
	AuthMethodSTSBound  = "sts-bound"
	AuthMethodSTSVendor = "sts-vendor"

	// Bound STS tokens are not refreshed in-process, so each one has to outlive the
	// process that fetched it: the mng process lives until its VM rotates (7 days),
	// the install process until its next scheduled restart (4 hours, plus 15 minutes
	// for jobs still draining at the boundary). A longer lifetime costs little:
	// the control plane expires a token once the process that owns it goes offline.
	MngTokenTTL     = 7 * 24 * time.Hour
	InstallTokenTTL = 4*time.Hour + 15*time.Minute
)

// TokenTTLForProcess returns the bound STS token lifetime a runner process should request.
func TokenTTLForProcess(process string) time.Duration {
	if process == "install" {
		return InstallTokenTTL
	}
	return MngTokenTTL
}

type fetchOptions struct {
	ttl time.Duration
}

type FetchOption func(*fetchOptions)

// WithTTL sets the token lifetime requested from endpoints that accept one (bound STS).
func WithTTL(ttl time.Duration) FetchOption {
	return func(o *fetchOptions) { o.ttl = ttl }
}

// FetchToken authenticates using cloud instance credentials and returns the token without writing to disk.
// It detects the cloud provider from the CLOUD_PROVIDER env var, falling back to auto-detection.
// For AWS, authMethod selects the authentication strategy: "" / "iid" (default) for Instance
// Identity Document (requires runnerID), "sts" for presigned STS and EC2 tags requests, or
// "sts-bound" for a bound STS request signed by the runner's own role (requires runnerID).
func FetchToken(ctx context.Context, apiClient nuonrunner.Client, authMethod, runnerID string, refreshable bool, opts ...FetchOption) (*FetchTokenResult, error) {
	options := fetchOptions{ttl: MngTokenTTL}
	for _, opt := range opts {
		opt(&options)
	}

	provider := detectCloudProvider(ctx)
	switch provider {
	case "gcp":
		return fetchTokenGCP(ctx, apiClient, runnerID, refreshable)
	case "aws":
		switch authMethod {
		case "sts":
			return fetchTokenAWSSTS(ctx, apiClient)
		case AuthMethodSTSBound:
			return fetchTokenAWSSTSBound(ctx, apiClient, runnerID, options.ttl)
		case AuthMethodSTSVendor:
			return nil, fmt.Errorf("auth method %q is not supported by this runner version", AuthMethodSTSVendor)
		default:
			return fetchTokenAWSIID(ctx, apiClient, runnerID)
		}
	default:
		return nil, fmt.Errorf("unsupported or undetected cloud provider: %s", provider)
	}
}

// FetchTokenAzure authenticates with Azure managed identity and returns the token without writing to disk.
func FetchTokenAzure(ctx context.Context, apiClient nuonrunner.Client, runnerID string, refreshable bool) (*FetchTokenResult, error) {
	f := &pkgazure.Fetcher{RunnerID: runnerID, Refreshable: refreshable}
	result, err := f.FetchToken(ctx, apiClient)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with azure: %w", err)
	}

	return &FetchTokenResult{
		RunnerID:  result.RunnerID,
		AccountID: result.AccountID,
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		Audience:  result.Audience,
	}, nil
}

// FetchAndStoreToken authenticates using cloud instance credentials and writes the token to disk.
func FetchAndStoreToken(ctx context.Context, apiClient nuonrunner.Client, authMethod, runnerID string) (*FetchTokenResult, error) {
	result, err := FetchToken(ctx, apiClient, authMethod, runnerID, false)
	if err != nil {
		return nil, err
	}

	return storeToken(result)
}

// FetchAndStoreTokenAzure authenticates with Azure managed identity and writes the token to disk.
func FetchAndStoreTokenAzure(ctx context.Context, apiClient nuonrunner.Client, runnerID string) (*FetchTokenResult, error) {
	result, err := FetchTokenAzure(ctx, apiClient, runnerID, false)
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

func runnerAPIHost() (string, error) {
	apiURL := os.Getenv("RUNNER_API_URL")
	if apiURL == "" {
		apiURL = "https://runner.nuon.co"
	}

	u, err := url.Parse(apiURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("RUNNER_API_URL %q has no host", apiURL)
	}
	return strings.ToLower(u.Host), nil
}

func fetchTokenAWSSTSBound(ctx context.Context, apiClient nuonrunner.Client, runnerID string, ttl time.Duration) (*FetchTokenResult, error) {
	if runnerID == "" {
		return nil, fmt.Errorf("runner ID is required for %s auth", AuthMethodSTSBound)
	}

	serverID, err := runnerAPIHost()
	if err != nil {
		return nil, err
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsRequest, err := pkgaws.PresignBoundSTS(ctx, cfg, serverID, runnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bound presigned STS request: %w", err)
	}

	req := pkgaws.BuildBoundSTSAuthRequest(AuthMethodSTSBound, runnerID, stsRequest)

	resp, err := apiClient.RunnerAuthAWSSTS(ctx, req, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with bound STS: %w", err)
	}
	if resp.Token == "" {
		return nil, fmt.Errorf("authentication failed: runner was not authenticated")
	}

	// ExpiresAt stays zero on purpose: the token is not refreshable, and a zero
	// expiry is what keeps auth.Refresher from starting.
	return &FetchTokenResult{
		RunnerID: resp.RunnerID,
		Token:    resp.Token,
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

func fetchTokenGCP(ctx context.Context, apiClient nuonrunner.Client, runnerID string, refreshable bool) (*FetchTokenResult, error) {
	apiURL := os.Getenv("RUNNER_API_URL")
	if apiURL == "" {
		apiURL = "https://runner.nuon.co"
	}

	if runnerID == "" {
		var err error
		runnerID, err = pkggcp.GetRunnerID(ctx)
		if err != nil {
			return nil, err
		}
	}

	identityToken, err := pkggcp.GetIdentityToken(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP identity token: %w", err)
	}

	resp, err := apiClient.RunnerAuthGCP(ctx, &models.ServiceRunnerAuthGCPRequest{
		IdentityToken: &identityToken,
		RunnerID:      runnerID,
		Refreshable:   refreshable,
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
		ExpiresAt:  tokenfetcher.ParseExpiresAt(resp.ExpiresAt),
	}, nil
}
