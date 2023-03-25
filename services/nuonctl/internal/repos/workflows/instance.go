package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/s3downloader"
)

// GetInstanceProvisionResponse returns a provision response
func (r *repo) GetInstanceProvisionRequest(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Request, error) {
	client, err := s3downloader.New(r.DeploymentsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	prefix := prefix.InstancePath(orgID,
		appID,
		componentID,
		deploymentID,
		installID,
	)
	key := filepath.Join(prefix, requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetInstanceProvisionResponse(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Response, error) {
	client, err := s3downloader.New(r.DeploymentsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	prefix := prefix.InstancePath(orgID,
		appID,
		componentID,
		deploymentID,
		installID,
	)
	key := filepath.Join(prefix, responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
