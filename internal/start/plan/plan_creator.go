package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-uploader"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
	shared "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/powertoolsdev/workers-deployments/internal/start/plan/config"
)

const (
	planFilename string = "plan.json"
)

type CreatePlanRequest struct {
	OrgID        string `json:"org_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	DeploymentID string `json:"deployment_id" validate:"required"`

	DeploymentsBucketPrefix        string `json:"deployments_bucket_prefix" validate:"required"`
	DeploymentsBucketAssumeRoleARN string `json:"deployments_bucket_assume_role_arn" validate:"required"`

	Component *componentv1.Component `validate:"required"`

	// NOTE(jm): while this isn't ideal, it saves a bunch of back and forth and managing var names in two places.
	// Given the nature of the plan itself, it's basically guaranteed that almost every value in the top level
	// config is needed here.
	//
	// Furthermore, given none of the values are "sensitive", it's much safer to pass this with the activity request
	// than have the activity struct itself manage a copy. By doing it this way, it versions with the workflow and
	// makes it easier for us to replace later on
	Config shared.Config `validate:"required"`
}

func (u CreatePlanRequest) validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type CreatePlanResponse struct {
	Plan *planv1.PlanRef
}

func (a *Activities) CreatePlan(ctx context.Context, req CreatePlanRequest) (CreatePlanResponse, error) {
	resp := CreatePlanResponse{}

	// create upload client
	assumeRoleOpt := uploader.WithAssumeRoleARN(req.DeploymentsBucketAssumeRoleARN)
	assumeRoleSessionOpt := uploader.WithAssumeSessionName("workers-deployments")
	uploadClient := uploader.NewS3Uploader(req.Config.DeploymentsBucket, req.DeploymentsBucketPrefix, assumeRoleOpt, assumeRoleSessionOpt)

	// fetch builder
	builder, err := a.planCreator.getConfigBuilder(req.Component)
	if err != nil {
		return resp, fmt.Errorf("unable to get config builder: %w", err)
	}

	// create plan + upload
	plan, err := a.planCreator.createPlan(req, builder)
	if err != nil {
		return resp, fmt.Errorf("unable to create plan: %w", err)
	}

	planRef, err := a.planCreator.uploadPlan(ctx, uploadClient, req, plan)
	if err != nil {
		return resp, fmt.Errorf("unable to upload plan: %w", err)
	}
	resp.Plan = planRef

	return resp, nil
}

type planCreator interface {
	getConfigBuilder(*componentv1.Component) (config.Builder, error)
	createPlan(CreatePlanRequest, config.Builder) (*planv1.BuildPlan, error)
	uploadPlan(context.Context, s3BlobUploader, CreatePlanRequest, *planv1.BuildPlan) (*planv1.PlanRef, error)
}

type planCreatorImpl struct{}

var _ planCreator = (*planCreatorImpl)(nil)

func (planCreatorImpl) getConfigBuilder(component *componentv1.Component) (config.Builder, error) {
	return config.NewStaticBuilder(), nil
}

func (planCreatorImpl) createPlan(req CreatePlanRequest, builder config.Builder) (*planv1.BuildPlan, error) {
	return nil, nil
}

func (planCreatorImpl) uploadPlan(ctx context.Context, uploader s3BlobUploader, req CreatePlanRequest, plan *planv1.BuildPlan) (*planv1.PlanRef, error) {
	byts, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := uploader.UploadBlob(ctx, byts, planFilename); err != nil {
		return nil, fmt.Errorf("unable to upload plan: %w", err)
	}

	return &planv1.PlanRef{
		Bucket:              req.Config.DeploymentsBucket,
		BucketKey:           filepath.Join(req.DeploymentsBucketPrefix, planFilename),
		BucketAssumeRoleArn: req.DeploymentsBucketAssumeRoleARN,
	}, nil
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}
