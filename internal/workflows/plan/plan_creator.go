package plan

import (
	"context"
	"fmt"

	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/powertoolsdev/workers-executors/internal/workflows/plan/config"
	"google.golang.org/protobuf/proto"
)

const (
	planFilename             string = "plan.json"
	defaultJobTimeoutSeconds uint64 = 3600
)

func (a *Activities) CreatePlanAct(ctx context.Context, req *planactivitiesv1.CreatePlanRequest) (*planactivitiesv1.CreatePlanResponse, error) {
	resp := &planactivitiesv1.CreatePlanResponse{}

	// create upload client
	//assumeRoleOpt := uploader.WithAssumeRoleARN(req.DeploymentsBucketAssumeRoleArn)
	//assumeRoleSessionOpt := uploader.WithAssumeSessionName("workers-deployments")
	//uploadClient := uploader.NewS3Uploader(req.Config.DeploymentsBucket, req.DeploymentsBucketPrefix, assumeRoleOpt, assumeRoleSessionOpt)

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

	planRef, err := a.planCreator.uploadPlan(ctx, nil, req, plan)
	if err != nil {
		return resp, fmt.Errorf("unable to upload plan: %w", err)
	}
	resp.Plan = planRef

	return resp, nil
}

type planCreator interface {
	getConfigBuilder(*componentv1.Component) (config.Builder, error)
	createPlan(*planactivitiesv1.CreatePlanRequest, config.Builder) (*planv1.WaypointPlan, error)
	uploadPlan(context.Context, s3BlobUploader, *planactivitiesv1.CreatePlanRequest, *planv1.WaypointPlan) (*planv1.PlanRef, error)
}

type planCreatorImpl struct{}

var _ planCreator = (*planCreatorImpl)(nil)

func (planCreatorImpl) getConfigBuilder(component *componentv1.Component) (config.Builder, error) {
	return config.NewStaticBuilder(), nil
}

func (planCreatorImpl) createPlan(req *planactivitiesv1.CreatePlanRequest, builder config.Builder) (*planv1.WaypointPlan, error) {
	//ecrRepoName := fmt.Sprintf("%s/%s", req.OrgId, req.AppId)
	//ecrRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", req.Config.OrgsECRRegistryID,
	//req.Config.OrgsECRRegion, ecrRepoName)

	//plan := &planv1.WaypointPlan{
	//Metadata: &planv1.Metadata{
	//OrgId:	     req.OrgId,
	//OrgShortId:	     req.OrgShortId,
	//AppId:	     req.AppId,
	//AppShortId:	     req.AppShortId,
	//DeploymentId:      req.DeploymentId,
	//DeploymentShortId: req.DeploymentShortId,
	//},
	//WaypointServer: &planv1.WaypointServerRef{
	//Address:		client.DefaultOrgServerAddress(req.Config.WaypointServerRootDomain, req.OrgID),
	//TokenSecretName:	fmt.Sprintf(req.Config.WaypointTokenSecretTemplate, req.OrgID),
	//TokenSecretNamespace: req.Config.WaypointTokenSecretNamespace,
	//},
	//EcrRepositoryRef: &planv1.ECRRepositoryRef{
	//RegistryId:	  req.Config.OrgsECRRegistryID,
	//RepositoryName: ecrRepoName,
	//RepositoryArn:  fmt.Sprintf("%s/%s", req.Config.OrgsECRRegistryARN, ecrRepoName),
	//RepositoryUri:  ecrRepoURI,
	//Tag:		  req.DeploymentID,
	//Region:	  req.Config.OrgsECRRegion,
	//},
	//WaypointRef: &planv1.WaypointRef{
	//Project:     req.OrgID,
	//Workspace:   req.AppID,
	//App:	       req.Component.Name,
	//SingletonId: fmt.Sprintf("%s-%s", req.DeploymentID, req.Component.Name),
	//Labels: map[string]string{
	//"deployment-id":  req.DeploymentID,
	//"app-id":	    req.AppID,
	//"org-id":	    req.OrgID,
	//"component-name": req.Component.Name,
	//},
	//RunnerId:		req.OrgID,
	//OnDemandRunnerConfig: req.OrgID,
	//JobTimeoutSeconds:	defaultJobTimeoutSeconds,
	//},
	//Outputs: &planv1.Outputs{
	//Bucket:	       req.Config.DeploymentsBucket,
	//BucketPrefix:        req.DeploymentsBucketPrefix,
	//BucketAssumeRoleArn: req.DeploymentsBucketAssumeRoleARN,

	//// TODO(jm): these aren't being used until we've fully implemented the executor
	//LogsKey:     filepath.Join(req.DeploymentsBucketPrefix, "logs.txt"),
	//EventsKey:   filepath.Join(req.DeploymentsBucketPrefix, "events.json"),
	//ArtifactKey: filepath.Join(req.DeploymentsBucketPrefix, "artifacts.json"),
	//},
	//Component: req.Component,
	//}

	//// configure builder
	//builder.WithComponent(req.Component)
	//builder.WithMetadata(plan.Metadata)
	//builder.WithECRRef(plan.EcrRepositoryRef)
	//cfg, cfgFmt, err := builder.Render()
	//if err != nil {
	//return nil, fmt.Errorf("unable to render config: %w", err)
	//}
	//plan.WaypointRef.HclConfig = string(cfg)
	//plan.WaypointRef.HclConfigFormat = cfgFmt.String()
	//return plan, nil
	return nil, nil
}

func (planCreatorImpl) uploadPlan(ctx context.Context, uploader s3BlobUploader, req *planactivitiesv1.CreatePlanRequest, plan *planv1.WaypointPlan) (*planv1.PlanRef, error) {
	byts, err := proto.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := uploader.UploadBlob(ctx, byts, planFilename); err != nil {
		return nil, fmt.Errorf("unable to upload plan: %w", err)
	}

	return &planv1.PlanRef{
		Bucket:              "",
		BucketKey:           "",
		BucketAssumeRoleArn: "",
	}, nil
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}
