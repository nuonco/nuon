package build

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
)

type BuildRequest struct {
	OrgID        string `json:"org_id" validate:"required"`
	AppID        string `json:"app_id" validate:"required"`
	DeploymentID string `json:"deployment_id" validate:"required"`

	Component waypoint.Component `json:"component" validate:"required"`
}

func (s BuildRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type BuildResponse struct {
	WorkflowIDs []string `json:"workflow_ids"`
}

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) Build(ctx workflow.Context, req BuildRequest) (BuildResponse, error) {
	resp := BuildResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(workers.Config{})

	if err := req.Validate(); err != nil {
		return resp, err
	}

	bucketPrefix := getS3Prefix(req)

	l.Debug("pre-config-generate")
	genReq := GenerateWaypointConfigRequest{
		BuildRequest: req,
		BucketName:   w.cfg.Bucket,
		BucketPrefix: bucketPrefix,
	}

	genResp, err := execGenerateWaypointConfig(ctx, act, genReq)
	if err != nil {
		return resp, err
	}
	l.Debug("generated config", "config", genResp)

	uwaReq := UpsertWaypointApplicationRequest{
		OrgID:     req.OrgID,
		Component: req.Component,

		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgID),
	}
	uwaResp, err := execUpsertWaypointApplication(ctx, act, uwaReq)
	l.Debug("upserted waypoint app", "response", uwaResp)
	if err != nil {
		return resp, err
	}

	qwjReq := QueueWaypointDeploymentJobRequest{
		OrgID:                req.OrgID,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgID),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		AppID:                req.AppID,
		DeploymentID:         req.DeploymentID,
		ComponentName:        req.Component.Name,
		ComponentType:        req.Component.Type,
	}
	qwjResp, err := execQueueWaypointDeploymentJob(ctx, act, qwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug(fmt.Sprintf("successfully upserted job for deployment: jobID = %s", qwjResp.JobID))

	vwjReq := ValidateWaypointDeploymentJobRequest{
		OrgID:                req.OrgID,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgID),
		JobID:                qwjResp.JobID,
	}
	vwjResp, err := execValidateWaypointDeploymentJob(ctx, act, vwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug(fmt.Sprintf("successfully validated job for deployment: %v", vwjResp))

	pwjReq := PollWaypointDeploymentJobRequest{
		OrgID:                req.OrgID,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgID),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		JobID:                qwjResp.JobID,
	}
	pwjResp, err := execPollWaypointDeploymentJob(ctx, act, pwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully polled deployment job: %s", qwjResp.JobID, pwjResp)

	uaReq := UploadArtifactRequest{
		OrgID:                req.OrgID,
		AppID:                req.AppID,
		ComponentName:        req.Component.Name,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgID),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		DeploymentID:         req.DeploymentID,
	}
	uaResp, err := execUploadArtifact(ctx, act, uaReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully polled deployment job:", uaResp)

	return resp, nil
}

func execGenerateWaypointConfig(
	ctx workflow.Context,
	act *Activities,
	req GenerateWaypointConfigRequest,
) (GenerateWaypointConfigResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := GenerateWaypointConfigResponse{}

	l.Debug("executing generate waypoint config activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.GenerateWaypointConfig, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func execUpsertWaypointApplication(
	ctx workflow.Context,
	act *Activities,
	req UpsertWaypointApplicationRequest,
) (UpsertWaypointApplicationResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := UpsertWaypointApplicationResponse{}

	l.Debug("executing upsert waypoint application", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.UpsertWaypointApplication, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func execQueueWaypointDeploymentJob(
	ctx workflow.Context,
	act *Activities,
	req QueueWaypointDeploymentJobRequest,
) (QueueWaypointDeploymentJobResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp QueueWaypointDeploymentJobResponse

	l.Debug("executing upsert waypoint application", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.QueueWaypointDeploymentJob, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execPollWaypointDeploymentJob(
	ctx workflow.Context,
	act *Activities,
	req PollWaypointDeploymentJobRequest,
) (PollWaypointDeploymentJobResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp PollWaypointDeploymentJobResponse

	l.Debug("executing poll waypoint deployment job", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.PollWaypointDeploymentJob, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execValidateWaypointDeploymentJob(
	ctx workflow.Context,
	act *Activities,
	req ValidateWaypointDeploymentJobRequest,
) (ValidateWaypointDeploymentJobResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp ValidateWaypointDeploymentJobResponse

	l.Debug("executing validate waypoint deployment job", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ValidateWaypointDeploymentJob, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execUploadArtifact(
	ctx workflow.Context,
	act *Activities,
	req UploadArtifactRequest,
) (UploadArtifactResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp UploadArtifactResponse

	l.Debug("executing upload artifact job", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.UploadArtifact, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
