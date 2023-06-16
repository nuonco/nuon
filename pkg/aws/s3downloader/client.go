package s3downloader

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

func (s *s3Downloader) getClient(ctx context.Context) (*s3.Client, error) {
	cfg, err := credentials.Fetch(ctx, s.Credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch credentials: %w", err)
	}

	return s3.NewFromConfig(cfg), nil
}
