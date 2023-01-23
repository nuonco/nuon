package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-uploader"
	"github.com/powertoolsdev/go-waypoint"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
	shared "github.com/powertoolsdev/workers-executors/internal"
	"github.com/powertoolsdev/workers-executors/internal/workflows/plan/config"
)

const (
	planFilename             string = "plan.json"
	defaultJobTimeoutSeconds uint64 = 3600
)

type CreatePlanRequest struct {
	OrgID        string `json:"org_id" validate:"required" faker:"len=26"`
	AppID        string `json:"app_id" validate:"required" faker:"len=26"`
	DeploymentID string `json:"deployment_id" validate:"required" faker:"len=26"`

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
	orgUUID, err := shortid.ToUUID(req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org ID: %w", err)
	}

	appUUID, err := shortid.ToUUID(req.AppID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	deploymentUUID, err := shortid.ToUUID(req.DeploymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid deployment ID: %w", err)
	}

	ecrRepoName := fmt.Sprintf("%s/%s", req.OrgID, req.AppID)
	ecrRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", req.Config.OrgsECRRegistryID,
		req.Config.OrgsECRRegion, ecrRepoName)

	plan := &planv1.BuildPlan{
		Metadata: &planv1.Metadata{
			OrgId:             orgUUID.String(),
			OrgShortId:        req.OrgID,
			AppId:             appUUID.String(),
			AppShortId:        req.AppID,
			DeploymentId:      deploymentUUID.String(),
			DeploymentShortId: req.DeploymentID,
		},
		WaypointServer: &planv1.WaypointServerRef{
			Address:              waypoint.DefaultOrgServerAddress(req.Config.WaypointServerRootDomain, req.OrgID),
			TokenSecretName:      fmt.Sprintf(req.Config.WaypointTokenSecretTemplate, req.OrgID),
			TokenSecretNamespace: req.Config.WaypointTokenSecretNamespace,
		},
		EcrRepositoryRef: &planv1.ECRRepositoryRef{
			RegistryId:     req.Config.OrgsECRRegistryID,
			RepositoryName: ecrRepoName,
			RepositoryArn:  fmt.Sprintf("%s/%s", req.Config.OrgsECRRegistryARN, ecrRepoName),
			RepositoryUri:  ecrRepoURI,
			Tag:            req.DeploymentID,
			Region:         req.Config.OrgsECRRegion,
		},
		WaypointBuild: &planv1.WaypointBuild{
			Project:     req.OrgID,
			Workspace:   req.AppID,
			App:         req.Component.Name,
			SingletonId: fmt.Sprintf("%s-%s", req.DeploymentID, req.Component.Name),
			Labels: map[string]string{
				"deployment-id":  req.DeploymentID,
				"app-id":         req.AppID,
				"org-id":         req.OrgID,
				"component-name": req.Component.Name,
			},
			RunnerId:             req.OrgID,
			OnDemandRunnerConfig: req.OrgID,
			JobTimeoutSeconds:    defaultJobTimeoutSeconds,
		},
		Outputs: &planv1.BuildOutputs{
			Bucket:              req.Config.DeploymentsBucket,
			BucketPrefix:        req.DeploymentsBucketPrefix,
			BucketAssumeRoleArn: req.DeploymentsBucketAssumeRoleARN,

			// TODO(jm): these aren't being used until we've fully implemented the executor
			LogsKey:     filepath.Join(req.DeploymentsBucketPrefix, "logs.txt"),
			EventsKey:   filepath.Join(req.DeploymentsBucketPrefix, "events.json"),
			ArtifactKey: filepath.Join(req.DeploymentsBucketPrefix, "artifacts.json"),
		},
		Component: req.Component,
	}

	// configure builder
	builder.WithComponent(req.Component)
	builder.WithMetadata(plan.Metadata)
	builder.WithECRRef(plan.EcrRepositoryRef)
	cfg, cfgFmt, err := builder.Render()
	if err != nil {
		return nil, fmt.Errorf("unable to render config: %w", err)
	}
	plan.WaypointBuild.HclConfig = string(cfg)
	plan.WaypointBuild.HclConfigFormat = cfgFmt.String()

	return plan, nil
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
