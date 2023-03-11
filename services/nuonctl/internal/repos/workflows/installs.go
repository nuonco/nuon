package workflows

import (
	"context"
	"fmt"

	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/s3downloader"
)

func (r *repo) GetInstallsResponse(ctx context.Context, key string) (*sharedv1.Response, error) {
	client, err := s3downloader.New(r.InstallsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}

func (r *repo) GetInstallsRequest(ctx context.Context, key string) (*sharedv1.Request, error) {
	client, err := s3downloader.New(r.InstallsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}
