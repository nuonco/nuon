package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-workflows-meta/prefix"
	"github.com/powertoolsdev/orgs-api/internal/downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

// GetInstanceProvisionRequest returns a provision request for an org
func (r *repo) GetInstanceProvisionRequest(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Request, error) {
	client, err := downloader.New(r.InstancesBucket.Name,
		downloader.WithAssumeRoleARN(r.InstancesBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.InstancesBucket.IamRoleSessionName))

	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstancePath(orgID, appID, componentID, deploymentID, installID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetInstanceProvisionResponse(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Response, error) {
	client, err := downloader.New(r.OrgsBucket.Name,
		downloader.WithAssumeRoleARN(r.OrgsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.OrgsBucket.IamRoleSessionName))

	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstancePath(orgID, appID, componentID, deploymentID, installID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
