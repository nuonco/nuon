package dal

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"google.golang.org/protobuf/proto"
)

func (r *client) GetInstanceSyncPlan(ctx context.Context, orgID, appID, componentID, deployID, installID string) (*planv1.Plan, error) {
	creds := r.deploymentsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.DeploymentsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	bucketKey := prefix.InstancePhasePath(orgID, appID, componentID, deployID, installID, "sync-container-image")
	plan, err := r.getPlan(ctx, client, bucketKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	return plan, nil
}

func (r *client) GetInstanceDeployPlan(ctx context.Context, orgID, appID, componentID, deployID, installID string) (*planv1.Plan, error) {
	creds := r.deploymentsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.DeploymentsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	bucketKey := prefix.InstancePhasePath(orgID, appID, componentID, deployID, installID, "deploy")
	plan, err := r.getPlan(ctx, client, bucketKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	return plan, nil
}

func (r *client) GetBuildPlan(ctx context.Context, orgID, appID, componentID, buildID string) (*planv1.Plan, error) {
	creds := r.deploymentsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.DeploymentsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	bucketKey := prefix.BuildPath(orgID, appID, componentID, buildID)
	plan, err := r.getPlan(ctx, client, bucketKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	return plan, nil
}

func (r *client) getPlan(ctx context.Context, client s3downloader.Downloader, key string) (*planv1.Plan, error) {
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	plan := &planv1.Plan{}
	if err := proto.Unmarshal(byts, plan); err != nil {
		return nil, fmt.Errorf("invalid plan: %w", err)
	}

	return plan, nil
}
