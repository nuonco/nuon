package fetchtoken

import (
	"context"
	"fmt"
	"os"
	"strings"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	pkgaws "github.com/nuonco/nuon/bins/runner/internal/pkg/aws"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
)

const (
	runnerAuthAWSMethodEnv          = "RUNNER_AUTH_AWS_METHOD"
	runnerAuthSPIFFEAudienceEnv     = "RUNNER_AUTH_SPIFFE_AUDIENCE"
	runnerAuthSPIFFESocketPathEnv   = "RUNNER_AUTH_SPIFFE_SOCKET_PATH"
	runnerAuthAWSPresignedMethod    = "presigned"
	runnerAuthAWSSPIFFEMethod       = "spiffe"
	defaultRunnerAuthSPIFFEAudience = "nuon-runner-auth-aws"
)

type FetchTokenResult struct {
	RunnerID   string
	InstanceID string
	AccountID  string
	TokenPath  string
}

func FetchAndStoreToken(ctx context.Context, apiClient nuonrunner.Client) (*FetchTokenResult, error) {
	authMethod := strings.ToLower(strings.TrimSpace(os.Getenv(runnerAuthAWSMethodEnv)))
	if authMethod == "" {
		authMethod = runnerAuthAWSPresignedMethod
	}

	var (
		resp *models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSResponse
		err  error
	)

	switch authMethod {
	case runnerAuthAWSSPIFFEMethod:
		resp, err = fetchTokenWithSPIFFE(ctx, apiClient)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with SPIFFE/SPIRE: %w", err)
		}
	default:
		resp, err = fetchTokenWithPresignedAWS(ctx, apiClient)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with AWS presigned requests: %w", err)
		}
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

func fetchTokenWithPresignedAWS(ctx context.Context, apiClient nuonrunner.Client) (*models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSResponse, error) {
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

	return resp, nil
}

func fetchTokenWithSPIFFE(ctx context.Context, apiClient nuonrunner.Client) (*models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSResponse, error) {
	audience := strings.TrimSpace(os.Getenv(runnerAuthSPIFFEAudienceEnv))
	if audience == "" {
		audience = defaultRunnerAuthSPIFFEAudience
	}

	options := []workloadapi.ClientOption{}
	if socketPath := strings.TrimSpace(os.Getenv(runnerAuthSPIFFESocketPathEnv)); socketPath != "" {
		options = append(options, workloadapi.WithAddr(socketPath))
	}

	jwtSVID, err := workloadapi.FetchJWTSVID(ctx, jwtsvid.Params{Audience: audience}, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SPIFFE JWT-SVID: %w", err)
	}

	req := &models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSRequest{
		SPIFFEJWTSVID: jwtSVID.Marshal(),
	}

	resp, err := apiClient.RunnerAuthAWS(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with SPIFFE JWT-SVID: %w", err)
	}

	return resp, nil
}
