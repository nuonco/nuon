package plan

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-uploader"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint"
	waypointbuild "github.com/powertoolsdev/workers-executors/internal/planners/waypoint/build"
	waypointsync "github.com/powertoolsdev/workers-executors/internal/planners/waypoint/sync"
	"google.golang.org/protobuf/proto"
)

const (
	planKey string = "plan.json"
)

func (a *Activities) CreatePlanAct(
	ctx context.Context,
	req *planactivitiesv1.CreatePlanRequest,
) (*planactivitiesv1.CreatePlanResponse, error) {
	resp := &planactivitiesv1.CreatePlanResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	planner, err := a.planCreator.getPlanner(req)
	if err != nil {
		return resp, fmt.Errorf("unable to get planner: %w", err)
	}

	planRef := &planv1.PlanRef{
		Bucket:              req.OrgMetadata.Buckets.DeploymentsBucket,
		BucketKey:           filepath.Join(planner.Prefix(), planKey),
		BucketAssumeRoleArn: req.OrgMetadata.IamRoleArns.DeploymentsRoleArn,
	}

	plan, err := planner.Plan(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get plan: %w", err)
	}

	// create upload client
	uploadClient, err := uploader.NewS3Uploader(a.v, uploader.WithBucketName(planRef.Bucket),
		uploader.WithAssumeSessionName("workers-executors"),
		uploader.WithAssumeRoleARN(planRef.BucketAssumeRoleArn))
	if err != nil {
		return resp, fmt.Errorf("unable to get uploader: %w", err)
	}

	err = a.planUploader.uploadPlan(ctx, uploadClient, planRef, plan)
	if err != nil {
		return resp, fmt.Errorf("unable to upload plan: %w", err)
	}
	resp.Plan = planRef

	return resp, nil
}

type planCreator interface {
	getPlanner(*planactivitiesv1.CreatePlanRequest) (planners.Planner, error)
}

type planCreatorImpl struct {
	v *validator.Validate
}

var _ planCreator = (*planCreatorImpl)(nil)

func (p *planCreatorImpl) getPlanner(req *planactivitiesv1.CreatePlanRequest) (planners.Planner, error) {
	waypointOpts := []waypoint.PlannerOption{
		waypoint.WithComponent(req.Component),
		waypoint.WithMetadata(req.Metadata),
		waypoint.WithOrgMetadata(req.OrgMetadata),
	}

	switch req.Type {
	case planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD:
		return waypointbuild.New(p.v, waypointOpts...)
	case planv1.PlanType_PLAN_TYPE_WAYPOINT_SYNC_IMAGE:
		return waypointsync.New(p.v, waypointOpts...)
	default:
		return nil, fmt.Errorf("unsupported plan type: %s", req.Type)
	}
}

type planUploader interface {
	uploadPlan(context.Context, s3BlobUploader, *planv1.PlanRef, *planv1.Plan) error
}

type planUploaderImpl struct{}

var _ planUploader = (*planUploaderImpl)(nil)

func (p *planUploaderImpl) uploadPlan(
	ctx context.Context,
	uploader s3BlobUploader,
	planRef *planv1.PlanRef,
	plan *planv1.Plan,
) error {
	byts, err := proto.Marshal(plan)
	if err != nil {
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := uploader.UploadBlob(ctx, byts, planRef.BucketKey); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}
