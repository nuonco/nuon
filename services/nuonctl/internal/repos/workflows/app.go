package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/pkg/workflows-meta/prefix"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/s3downloader"
)

// GetAppProvisionRequest returns a provision request for an org
func (r *repo) GetAppProvisionRequest(ctx context.Context, orgID, appID string) (*sharedv1.Request, error) {
	client, err := s3downloader.New(r.AppsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.AppPath(orgID, appID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetAppProvisionResponse(ctx context.Context, orgID, appID string) (*sharedv1.Response, error) {
	client, err := s3downloader.New(r.AppsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.AppPath(orgID, appID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
