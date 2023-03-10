package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-workflows-meta/prefix"
	"github.com/powertoolsdev/orgs-api/internal/downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

// GetAppProvisionRequest returns a provision request for an org
func (r *repo) GetAppProvisionRequest(ctx context.Context, orgID, appID string) (*sharedv1.Request, error) {
	client, err := downloader.New(r.AppsBucket.Name,
		downloader.WithAssumeRoleARN(r.AppsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.AppsBucket.IamRoleSessionName))

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
	client, err := downloader.New(r.OrgsBucket.Name,
		downloader.WithAssumeRoleARN(r.OrgsBucket.IamRoleArn),
		downloader.WithAssumeRoleSessionName(r.OrgsBucket.IamRoleSessionName))

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
