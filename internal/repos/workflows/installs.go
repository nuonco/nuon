package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-workflows-meta/prefix"
	"github.com/powertoolsdev/orgs-api/internal/downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

// GetInstallProvisionRequest returns a provision request for an org
func (r *repo) GetInstallProvisionRequest(ctx context.Context, orgID, appID, installID string) (*sharedv1.Request, error) {
	client, err := downloader.New(r.InstallsBucket.Name,
		downloader.WithAssumeRoleARN(r.InstallsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.InstallsBucket.IamRoleSessionName))

	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstallPath(orgID, appID, installID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetInstallProvisionResponse(ctx context.Context, orgID, appID, installID string) (*sharedv1.Response, error) {
	client, err := downloader.New(r.OrgsBucket.Name,
		downloader.WithAssumeRoleARN(r.OrgsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.OrgsBucket.IamRoleSessionName))

	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstallPath(orgID, appID, installID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
