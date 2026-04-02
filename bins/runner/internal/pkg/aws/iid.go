package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// IIDResult holds the instance identity document and its RSA-2048 signature.
type IIDResult struct {
	Document  string
	Signature string
}

// GetInstanceIdentityDocument fetches the IID and RSA-2048 signature from
// IMDSv2.
func GetInstanceIdentityDocument(ctx context.Context) (*IIDResult, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := imds.NewFromConfig(cfg)

	docOutput, err := client.GetDynamicData(ctx, &imds.GetDynamicDataInput{
		Path: "instance-identity/document",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance identity document: %w", err)
	}
	defer docOutput.Content.Close()

	docBytes, err := io.ReadAll(docOutput.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read identity document: %w", err)
	}

	sigOutput, err := client.GetDynamicData(ctx, &imds.GetDynamicDataInput{
		Path: "instance-identity/rsa2048",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get RSA-2048 signature: %w", err)
	}
	defer sigOutput.Content.Close()

	sigBytes, err := io.ReadAll(sigOutput.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSA-2048 signature: %w", err)
	}

	sigStr := strings.TrimSpace(string(sigBytes))
	sigStr = strings.ReplaceAll(sigStr, "\n", "")
	sigStr = strings.ReplaceAll(sigStr, "\r", "")

	if _, err := base64.StdEncoding.DecodeString(sigStr); err != nil {
		return nil, fmt.Errorf("IMDS returned invalid base64 signature: %w", err)
	}

	return &IIDResult{
		Document:  string(docBytes),
		Signature: sigStr,
	}, nil
}
