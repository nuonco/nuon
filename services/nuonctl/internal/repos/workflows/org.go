package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/workflows-meta/prefix"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/s3downloader"
	sharedv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/shared/v1"
)

// GetOrgProvisionRequest returns a provision request for an org
func (r *repo) GetOrgProvisionRequest(ctx context.Context, orgID string) (*sharedv1.Request, error) {
	client, err := s3downloader.New(r.OrgsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.OrgPath(orgID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetOrgProvisionResponse(ctx context.Context, orgID string) (*sharedv1.Response, error) {
	client, err := s3downloader.New(r.OrgsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.OrgPath(orgID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
