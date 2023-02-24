package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/orgs-api/internal/downloader"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

// GetAppProvisionRequest returns a provision request for an org
func (r *repo) GetAppProvisionRequest(ctx context.Context) (*sharedv1.Request, error) {
	orgCtx, err := r.ctxGetter(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get context: %w", err)
	}
	bucket, ok := orgCtx.Buckets[orgcontext.BucketTypeOrgs]
	if !ok {
		return nil, errBucketNotFound
	}

	client, err := downloader.New(bucket.Name, downloader.WithAssumeRoleARN(bucket.AssumeRoleARN), downloader.WithAssumeRoleSessionName(bucket.AssumeRoleName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	path := filepath.Join(orgCtx.Buckets[orgcontext.BucketTypeOrgs].Prefix, requestFilename)
	byts, err := client.GetBlob(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	var resp sharedv1.Request
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &resp, nil
}

func (r *repo) GetAppProvisionResponse(ctx context.Context) (*sharedv1.Response, error) {
	orgCtx, err := r.ctxGetter(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get context: %w", err)
	}
	bucket, ok := orgCtx.Buckets[orgcontext.BucketTypeOrgs]
	if !ok {
		return nil, errBucketNotFound
	}

	client, err := downloader.New(bucket.Name, downloader.WithAssumeRoleARN(bucket.AssumeRoleARN), downloader.WithAssumeRoleSessionName(bucket.AssumeRoleName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	path := filepath.Join(orgCtx.Buckets[orgcontext.BucketTypeOrgs].Prefix, responseFilename)
	byts, err := client.GetBlob(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	return unmarshalResponse(byts)
}
