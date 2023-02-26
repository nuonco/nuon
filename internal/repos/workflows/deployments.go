package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-workflows-meta/prefix"
	"github.com/powertoolsdev/orgs-api/internal/downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

// GetDeploymentProvisionRequest returns a provision request for an org
func (r *repo) GetDeploymentProvisionRequest(ctx context.Context, orgID, appID, componentID, deploymentID string) (*sharedv1.Request, error) {
	client, err := downloader.New(r.DeploymentsBucket.Name,
		downloader.WithAssumeRoleARN(r.DeploymentsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.DeploymentsBucket.IamRoleSessionName))

	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.DeploymentPath(orgID, appID, componentID, deploymentID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetDeploymentProvisionResponse(ctx context.Context, orgID, appID, componentID, deploymentID string) (*sharedv1.Response, error) {
	client, err := downloader.New(r.OrgsBucket.Name,
		downloader.WithAssumeRoleARN(r.OrgsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.OrgsBucket.IamRoleSessionName))

	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.DeploymentPath(orgID, appID, componentID, deploymentID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
