package workflows

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/s3downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func (r *repo) GetDeploymentsResponse(ctx context.Context, key string) (*sharedv1.Response, error) {
	client, err := s3downloader.New(r.DeploymentsBucket,
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

func (r *repo) GetDeploymentsRequest(ctx context.Context, key string) (*sharedv1.Request, error) {
	client, err := s3downloader.New(r.DeploymentsBucket,
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
