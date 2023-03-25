package s3downloader

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

func (s *s3Downloader) getClient(ctx context.Context) (*s3.Client, error) {
	assumer, err := assumerole.New(s.v, assumerole.WithRoleARN(s.AssumeRoleARN), assumerole.WithRoleSessionName(s.AssumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to create role assumer: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to assume role: %w", err)
	}

	return s3.NewFromConfig(cfg), nil
}
