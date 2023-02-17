package workflows

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-workflows-meta/prefix"
	"github.com/powertoolsdev/nuonctl/internal/s3downloader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func EnsureShortIDs(ids ...string) ([]string, error) {
	shortIDs := make([]string, len(ids))

	for idx, id := range ids {
		_, err := shortid.ToUUID(id)
		if err == nil {
			shortIDs[idx] = id
			continue
		}

		shortID, err := shortid.ParseString(id)
		if err != nil {
			return nil, fmt.Errorf("unable to parse id %d %s into shortid: %w", idx, id, err)
		}
		shortIDs[idx] = shortID
	}

	return shortIDs, nil
}

// GetInstanceProvisionResponse returns a provision response
func (r *repo) GetInstanceProvisionRequest(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Request, error) {
	ids, err := EnsureShortIDs(orgID, appID, componentID, deploymentID, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to ensure ids: %w", err)
	}

	client, err := s3downloader.New(r.DeploymentsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstancePath(ids[0], ids[1], ids[2], ids[3], ids[4]), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalRequest(byts)
}

func (r *repo) GetInstanceProvisionResponse(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Response, error) {
	ids, err := EnsureShortIDs(orgID, appID, componentID, deploymentID, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to ensure ids: %w", err)
	}

	client, err := s3downloader.New(r.OrgsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstancePath(ids[0], ids[1], ids[2], ids[3], ids[4]), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
