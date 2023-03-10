package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-workflows-meta/prefix"
	"github.com/powertoolsdev/nuonctl/internal/s3downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
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
