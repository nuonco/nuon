package aws

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

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
)

// getRegionFromIMDS fetches the current region from EC2 instance metadata.
func getRegionFromIMDS(ctx context.Context) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	imdsClient := imds.NewFromConfig(cfg)
	regionOutput, err := imdsClient.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get region from IMDS: %w", err)
	}

	return regionOutput.Region, nil
}

// GetPresignedSTSRequest creates a presigned STS GetCallerIdentity request.
// The presigned request can be sent to another service which will make the actual
// STS call to validate the caller's identity.
func GetPresignedSTSRequest(ctx context.Context) (*awstypes.PresignedRequest, error) {
	region, err := getRegionFromIMDS(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	presignClient := sts.NewPresignClient(stsClient)

	presignedReq, err := presignClient.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to presign GetCallerIdentity: %w", err)
	}

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

// GetPresignedInstanceTagsRequest creates a presigned EC2 DescribeTags request
// for the current instance. This allows another service to fetch instance tags
// to verify the runner's identity.
func GetPresignedInstanceTagsRequest(ctx context.Context) (*awstypes.PresignedRequest, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

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

	regionOutput, err := imdsClient.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get region from IMDS: %w", err)
	}
	region := regionOutput.Region

	return presignEC2DescribeTags(ctx, cfg, region, instanceID)
}

// presignEC2DescribeTags manually creates and signs an EC2 DescribeTags request
func presignEC2DescribeTags(ctx context.Context, cfg aws.Config, region, instanceID string) (*awstypes.PresignedRequest, error) {
	params := url.Values{}
	params.Set("Action", "DescribeTags")
	params.Set("Version", "2016-11-15")
	params.Set("Filter.1.Name", "resource-id")
	params.Set("Filter.1.Value.1", instanceID)

	endpoint := fmt.Sprintf("https://ec2.%s.amazonaws.com/", region)
	reqURL := endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	signer := v4.NewSigner()
	payloadHash := sha256.Sum256([]byte{})
	payloadHashHex := hex.EncodeToString(payloadHash[:])

	err = signer.SignHTTP(ctx, creds, req, payloadHashHex, "ec2", region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

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
