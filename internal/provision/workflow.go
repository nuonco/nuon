package provision

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-instances/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
	defaultDeployTimeout   = time.Minute * 15
)

type ProvisionRequest struct {
	OrgID        string             `json:"org_id" validate:"required"`
	AppID        string             `json:"app_id" validate:"required"`
	DeploymentID string             `json:"deployment_id" validate:"required"`
	InstallID    string             `json:"install_id" validate:"required,min=1"`
	Component    waypoint.Component `json:"component" validate:"required"`
}

func (s ProvisionRequest) validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type ProvisionResponse struct{}

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w *Workflow) Provision(ctx workflow.Context, req ProvisionRequest) (ProvisionResponse, error) {
	resp := ProvisionResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(nil)

	l.Debug("pre-validate request")
	if err := req.validate(); err != nil {
		l.Error(err.Error())
		return resp, err
	}

	bucketPrefix := getS3Prefix(req)

	uploadReq := UploadMetadataRequest{
		Info:         *workflow.GetInfo(ctx),
		BucketName:   w.cfg.Bucket,
		BucketPrefix: bucketPrefix,
	}
	uploadResp, err := execUploadMetadata(ctx, act, uploadReq)
	if err != nil {
		return resp, err
	}

	l.Debug("uploaded workflow metadata", "reponse", uploadResp)

	uwaReq := UpsertWaypointApplicationRequest{
		OrgID:        req.OrgID,
		AppID:        req.AppID,
		InstallID:    req.InstallID,
		DeploymentID: req.DeploymentID,
		Component:    req.Component,

		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID),
	}
	uwaResp, err := execUpsertWaypointApplication(ctx, act, uwaReq)
	l.Debug("upserted waypoint app", uwaResp)
	if err != nil {
		return resp, err
	}

	qwjReq := QueueWaypointDeploymentJobRequest{
		OrgID:                req.OrgID,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		AppID:                req.AppID,
		InstallID:            req.InstallID,
		DeploymentID:         req.DeploymentID,
		ComponentName:        req.Component.Name,
	}
	qwjResp, err := execQueueWaypointDeploymentJob(ctx, act, qwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully upserted job for deployment: jobID", qwjResp.JobID)

	pwjReq := PollWaypointDeploymentJobRequest{
		OrgID:                req.OrgID,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		JobID:                qwjResp.JobID,
	}
	pwjResp, err := execPollWaypointDeploymentJob(ctx, act, pwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully polled deployment job: ", qwjResp.JobID, pwjResp)

	shnReq := SendHostnameNotificationRequest{
		OrgID:                req.OrgID,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID),
		InstallID:            req.InstallID,
		AppID:                req.AppID,
	}
	shnResp, err := execSendHostnameNotification(ctx, act, shnReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully sent hostname notification: ", shnResp)

	l.Debug("successfully provisioned instance of deployment ", req.DeploymentID, " on installation ", req.InstallID)
	return resp, nil
}

func execUpsertWaypointApplication(
	ctx workflow.Context,
	act *Activities,
	req UpsertWaypointApplicationRequest,
) (GenerateWaypointConfigResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := GenerateWaypointConfigResponse{}

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

	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultDeployTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	l.Debug("executing poll waypoint deployment job", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.PollWaypointDeploymentJob, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execUploadMetadata(
	ctx workflow.Context,
	act *Activities,
	req UploadMetadataRequest,
) (UploadResultResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp UploadResultResponse

	l.Debug("executing upload metadata job", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.UploadMetadata, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execSendHostnameNotification(
	ctx workflow.Context,
	act *Activities,
	req SendHostnameNotificationRequest,
) (SendHostnameNotificationResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp SendHostnameNotificationResponse

	l.Debug("executing send hostname notification", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.SendHostnameNotification, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
