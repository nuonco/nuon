package plan

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/go-uploader"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners/terraform/sandbox"
	"go.temporal.io/sdk/activity"
)

//nolint:unparam // NOTE(jdt): trying to keep these methods consistent, hence the unused error output
func (w *wkflow) sandboxPlanRequest(typ planv1.PlanType, req *planv1.Sandbox) (*planactivitiesv1.CreateSandboxPlan, error) {
	plan := &planactivitiesv1.CreateSandboxPlan{
		Type:    typ,
		Sandbox: req,
		Module: &planactivitiesv1.Bucket{
			Name:   w.cfg.SandboxBucket,
			Region: w.cfg.SandboxBucketRegion,
			AssumeRoleDetails: &planv1.AssumeRoleDetails{
				AssumeArn: fmt.Sprintf(w.cfg.OrgsInstallationsRoleTemplate, req.OrgId),
			},
		},
		Backend: &planactivitiesv1.Bucket{
			Name:   w.cfg.InstallationsBucket,
			Region: w.cfg.InstallationsBucketRegion,
		},

		Plan: &planactivitiesv1.Bucket{
			Name:   w.cfg.InstallationsBucket,
			Region: w.cfg.InstallationsBucketRegion,
			AssumeRoleDetails: &planv1.AssumeRoleDetails{
				AssumeArn: fmt.Sprintf(w.cfg.OrgsInstallationsRoleTemplate, req.OrgId),
			},
		},
	}
	return plan, nil
}

func (a *Activities) CreateTerraformSandboxPlan(
	ctx context.Context,
	req *planactivitiesv1.CreateSandboxPlan,
) (*planactivitiesv1.CreatePlanResponse, error) {
	resp := &planactivitiesv1.CreatePlanResponse{}
	l := activity.GetLogger(ctx)

	l.Debug("starting create terraform sandbox plan")
	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	l.Debug("creating sandbox planner")
	planner, err := sandbox.New(
		a.v,
		sandbox.WithPlan(req),
		sandbox.WithLogger(l),
	)
	if err != nil {
		return resp, fmt.Errorf("unable to get planner: %w", err)
	}

	l.Debug("creating sandbox plan")
	plan, err := planner.Plan(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get plan: %w", err)
	}

	planRef := &planv1.PlanRef{
		Bucket:              req.Plan.Name,
		BucketKey:           filepath.Join(planner.Prefix(), planKey),
		BucketAssumeRoleArn: req.Plan.AssumeRoleDetails.AssumeArn,
	}

	l.Debug("uploading sandbox plan")
	uploadClient, err := uploader.NewS3Uploader(
		a.v,
		uploader.WithBucketName(planRef.Bucket),
		uploader.WithAssumeRoleARN(planRef.BucketAssumeRoleArn),
		uploader.WithAssumeSessionName("workers-executors"),
	)
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
