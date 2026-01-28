package fetchtoken

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
	awstypes "github.com/nuonco/nuon/pkg/types/aws"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
		Status: models.AppRunnerJobStatusFinished,
	}); err != nil {
		return err
	}
	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("exec", zap.String("job_type", "fetch_token"))

	// Get presigned STS request for identity verification
	stsRequest, err := h.getPresignedSTSRequest(ctx)
	if err != nil {
		return fmt.Errorf("failed to get presigned STS request: %w", err)
	}

	// Get presigned EC2 tags request for runner ID verification
	tagsRequest, err := h.getPresignedInstanceTagsRequest(ctx)
	if err != nil {
		return fmt.Errorf("failed to get presigned EC2 tags request: %w", err)
	}

	// Convert to SDK model types
	req := &models.ServiceRunnerAuthAWSRequest{
		Sts: struct{ models.AwsPresignedRequest }{
			AwsPresignedRequest: models.AwsPresignedRequest{
				Method:  stsRequest.Method,
				URL:     stsRequest.URL,
				Headers: stsRequest.Headers,
			},
		},
		Tags: struct{ models.AwsPresignedRequest }{
			AwsPresignedRequest: models.AwsPresignedRequest{
				Method:  tagsRequest.Method,
				URL:     tagsRequest.URL,
				Headers: tagsRequest.Headers,
			},
		},
	}

	// Call the API to authenticate and get a token
	l.Info("authenticating with AWS presigned requests")
	resp, err := h.apiClient.RunnerAuthAWS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to authenticate with AWS: %w", err)
	}

	if !resp.Authenticated {
		return fmt.Errorf("authentication failed: runner was not authenticated")
	}

	l.Info("authentication successful",
		zap.String("runner_id", resp.RunnerID),
		zap.String("instance_id", resp.InstanceID),
		zap.String("aws_account_id", resp.AccountID))

	// Write the token to the token file
	if err := h.writeToken(resp.Token); err != nil {
		return fmt.Errorf("failed to write token: %w", err)
	}

	l.Info("token written successfully", zap.String("path", monitor.RunnerTokenFilename))

	// Mark the job as finished
	if err := h.finishJob(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("failed to finish job: %w", err)
	}

	return nil
}

func (h *handler) writeToken(token string) error {
	// this just wraps the method from pkg/monitor
	err := monitor.WriteRunnerTokenFile(token)
	if err != nil {
		return err
	}
	return nil
}

// getPresignedSTSRequest creates a presigned STS GetCallerIdentity request.
// The presigned request can be sent to another service which will make the actual
// STS call to validate the caller's identity.
func (h *handler) getPresignedSTSRequest(ctx context.Context) (*awstypes.PresignedRequest, error) {
	// Load AWS config from instance credentials (IMDS)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create STS presign client
	stsClient := sts.NewFromConfig(cfg)
	presignClient := sts.NewPresignClient(stsClient)

	// Create presigned GetCallerIdentity request
	presignedReq, err := presignClient.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to presign GetCallerIdentity: %w", err)
	}

	// Extract headers from the presigned request
	headers := make(map[string]string)
	for key, values := range presignedReq.SignedHeader {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &awstypes.PresignedRequest{
		Method:  presignedReq.Method,
		URL:     presignedReq.URL,
		Headers: headers,
	}, nil
}

// getPresignedInstanceTagsRequest creates a presigned EC2 DescribeTags request
// for the current instance. This allows another service to fetch instance tags
// to verify the runner's identity.
func (h *handler) getPresignedInstanceTagsRequest(ctx context.Context) (*awstypes.PresignedRequest, error) {
	// Load AWS config from instance credentials (IMDS)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 1. Get instance ID from IMDS
	imdsClient := imds.NewFromConfig(cfg)
	instanceIDOutput, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "instance-id",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance ID from IMDS: %w", err)
	}
	defer instanceIDOutput.Content.Close()

	instanceIDBytes := make([]byte, 64)
	n, err := instanceIDOutput.Content.Read(instanceIDBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read instance ID: %w", err)
	}
	instanceID := strings.TrimSpace(string(instanceIDBytes[:n]))

	// Get region from IMDS
	regionOutput, err := imdsClient.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get region from IMDS: %w", err)
	}
	region := regionOutput.Region

	// 2. Create presigned EC2 DescribeTags request
	// EC2 doesn't have a built-in presign client, so we manually construct and sign the request
	return h.presignEC2DescribeTags(ctx, cfg, region, instanceID)
}

// presignEC2DescribeTags manually creates and signs an EC2 DescribeTags request
func (h *handler) presignEC2DescribeTags(ctx context.Context, cfg aws.Config, region, instanceID string) (*awstypes.PresignedRequest, error) {
	// Build the EC2 DescribeTags query parameters
	params := url.Values{}
	params.Set("Action", "DescribeTags")
	params.Set("Version", "2016-11-15")
	params.Set("Filter.1.Name", "resource-id")
	params.Set("Filter.1.Value.1", instanceID)

	endpoint := fmt.Sprintf("https://ec2.%s.amazonaws.com/", region)
	reqURL := endpoint + "?" + params.Encode()

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get credentials for signing
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	// Sign the request using AWS Signature Version 4
	signer := v4.NewSigner()
	payloadHash := sha256.Sum256([]byte{}) // Empty payload for GET request
	payloadHashHex := hex.EncodeToString(payloadHash[:])

	err = signer.SignHTTP(ctx, creds, req, payloadHashHex, "ec2", region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Extract headers
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &awstypes.PresignedRequest{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: headers,
	}, nil
}
