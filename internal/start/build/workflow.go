package build

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-waypoint"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
	defaultBuildTimeout    = time.Minute * 60
)

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

func (w *wkflow) Build(ctx workflow.Context, req *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
	resp := &buildv1.BuildResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(workers.Config{})

	// TODO(jm): this will go away once this workflow leverages using an actual plan, it's just hardcoded cruft for
	// now to keep things working until that world "exists".
	component := waypoint.Component{
		Type:              "public",
		ContainerImageURL: "kennethreitz/httpbin",
		Name:              "httpbin",
	}

	if err := req.Validate(); err != nil {
		return resp, err
	}

	bucketPrefix := getS3Prefix(req)

	uwaReq := UpsertWaypointApplicationRequest{
		OrgID:                req.OrgId,
		Component:            component,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgId),
	}
	uwaResp, err := execUpsertWaypointApplication(ctx, act, uwaReq)
	l.Debug("upserted waypoint app", "response", uwaResp)
	if err != nil {
		return resp, err
	}

	qwjReq := QueueWaypointDeploymentJobRequest{
		OrgID:                req.OrgId,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgId),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		AppID:                req.AppId,
		DeploymentID:         req.DeploymentId,
		ComponentName:        component.Name,
		ComponentType:        component.Type,
	}
	qwjResp, err := execQueueWaypointDeploymentJob(ctx, act, qwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug(fmt.Sprintf("successfully upserted job for deployment: jobID = %s", qwjResp.JobID))

	vwjReq := ValidateWaypointDeploymentJobRequest{
		OrgID:                req.OrgId,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgId),
		JobID:                qwjResp.JobID,
	}
	vwjResp, err := execValidateWaypointDeploymentJob(ctx, act, vwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug(fmt.Sprintf("successfully validated job for deployment: %v", vwjResp))

	pwjReq := PollWaypointBuildJobRequest{
		OrgID:                req.OrgId,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgId),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		JobID:                qwjResp.JobID,
	}
	pwjResp, err := execPollWaypointBuildJob(ctx, act, pwjReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully polled deployment job: %s", qwjResp.JobID, pwjResp)

	uaReq := UploadArtifactRequest{
		OrgID:                req.OrgId,
		AppID:                req.AppId,
		ComponentName:        component.Name,
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointOrgServerRootDomain, req.OrgId),
		BucketName:           w.cfg.Bucket,
		BucketPrefix:         bucketPrefix,
		DeploymentID:         req.DeploymentId,
	}
	uaResp, err := execUploadArtifact(ctx, act, uaReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully uploaded deployment artifact:", uaResp)

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

func execPollWaypointBuildJob(
	ctx workflow.Context,
	act *Activities,
	req PollWaypointBuildJobRequest,
) (PollWaypointBuildJobResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp PollWaypointBuildJobResponse

	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultBuildTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	l.Debug("executing poll waypoint deployment job", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.PollWaypointBuildJob, req)
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
