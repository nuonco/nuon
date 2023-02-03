package plan

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-uploader"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint"
	waypointbuild "github.com/powertoolsdev/workers-executors/internal/planners/waypoint/build"
	"google.golang.org/protobuf/proto"
)

const (
	planFilename             string = "plan.json"
	defaultJobTimeoutSeconds uint64 = 3600
)

func (a *Activities) CreatePlanAct(ctx context.Context, req *planactivitiesv1.CreatePlanRequest) (*planactivitiesv1.CreatePlanResponse, error) {
	resp := &planactivitiesv1.CreatePlanResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	planner, err := a.planCreator.getPlanner(req)
	if err != nil {
		return resp, fmt.Errorf("unable to get planner: %w", err)
	}

	planRef := planner.GetPlanRef()

	plan, err := planner.GetPlan(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get plan: %w", err)
	}

	// create upload client
	assumeRoleOpt := uploader.WithAssumeRoleARN(planRef.BucketAssumeRoleArn)
	assumeRoleSessionOpt := uploader.WithAssumeSessionName("workers-deployments")
	uploadClient := uploader.NewS3Uploader(planRef.Bucket, planRef.BucketKey, assumeRoleOpt, assumeRoleSessionOpt)

	err = a.planCreator.uploadPlan(ctx, uploadClient, planRef, plan)
	if err != nil {
		return resp, fmt.Errorf("unable to upload plan: %w", err)
	}
	resp.Plan = planRef

	return resp, nil
}

type planCreator interface {
	getPlanner(*planactivitiesv1.CreatePlanRequest) (planners.Planner, error)
	uploadPlan(context.Context, s3BlobUploader, *planv1.PlanRef, *planv1.WaypointPlan) error
}

type planCreatorImpl struct {
	v *validator.Validate
}

var _ planCreator = (*planCreatorImpl)(nil)

func (p *planCreatorImpl) getPlanner(req *planactivitiesv1.CreatePlanRequest) (planners.Planner, error) {
	var (
		err     error
		planner planners.Planner
	)

	switch req.Type {
	case planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD:
		planner, err = waypointbuild.New(p.v,
			waypoint.WithComponent(req.Component),
			waypoint.WithMetadata(req.Metadata),
			waypoint.WithOrgMetadata(req.OrgMetadata),
		)
	default:
		return nil, fmt.Errorf("unsupported plan type: %s", req.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to get %s planner", req.Type)
	}

	return planner, nil
}

func (p *planCreatorImpl) uploadPlan(ctx context.Context, uploader s3BlobUploader, planRef *planv1.PlanRef, plan *planv1.WaypointPlan) error {
	byts, err := proto.Marshal(plan)
	if err != nil {
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := uploader.UploadBlob(ctx, byts, planFilename); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}
