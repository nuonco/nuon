package deprovision

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/shortid"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"go.temporal.io/sdk/workflow"
)

// DeprovisionRequest includes the set of arguments needed to deprovision a sandbox
type DeprovisionRequest struct {
	OrgID     string `json:"org_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	InstallID string `json:"install_id" validate:"required"`

	SandboxSettings struct {
		Name    string `json:"name" validate:"required"`
		Version string `json:"version" validate:"required"`
	} `json:"sandbox_settings" validate:"required"`

	// NOTE(jm): the following are customer provided values, consider adding a customer or external prefix?
	AwsRegion     string `json:"aws_region" validate:"required"`
	AssumeRoleArn string `json:"assume_role_arn" validate:"required"`
}

func (d DeprovisionRequest) validate() error {
	validate := validator.New()
	return validate.Struct(d)
}

type DeprovisionResponse struct{}

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

func (w wkflow) finishWithErr(ctx workflow.Context, req DeprovisionRequest, act *Activities, step string, err error) {
	l := workflow.GetLogger(ctx)
	finishReq := FinishRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationStateBucket,
		Success:             false,
		ErrorStep:           step,
		ErrorMessage:        fmt.Sprintf("%s", err),
	}

	if resp, execErr := execFinish(ctx, act, finishReq); execErr != nil {
		l.Debug("unable to finish with error: %w", execErr, resp)
	}
}

// Deprovision method destroys the infrastructure for an installation
func (w wkflow) Deprovision(ctx workflow.Context, req DeprovisionRequest) (DeprovisionResponse, error) {
	var resp DeprovisionResponse
	l := workflow.GetLogger(ctx)

	l.Debug("validating deprovision request")
	if err := req.validate(); err != nil {
		l.Debug("unable to validate terraform destroy request: %w", err)
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	// parse IDs into short IDs, and use them for all subsequent requests
	orgID, err := shortid.ParseString(req.OrgID)
	if err != nil {
		return resp, fmt.Errorf("unable to get short org ID: %w", err)
	}
	appID, err := shortid.ParseString(req.AppID)
	if err != nil {
		return resp, fmt.Errorf("unable to get short org ID: %w", err)
	}
	installID, err := shortid.ParseString(req.InstallID)
	if err != nil {
		return resp, fmt.Errorf("unable to get short install ID: %w", err)
	}

	// NOTE(jm): set the ids to short ids on the request, so every other part of this workflow uses shortids
	req.AppID = appID
	req.OrgID = orgID
	req.InstallID = installID

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	act := NewActivities(nil)

	stReq := StartRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationStateBucket,
	}
	_, err = execStart(ctx, act, stReq)
	if err != nil {
		l.Debug("unable to execute start activity: %w", err)
		return resp, fmt.Errorf("unable to execute start activity: %w", err)
	}

	dtReq := DestroyTerraformRequest{
		DeprovisionRequest: req,

		InstallationStateBucketName:   w.cfg.InstallationStateBucket,
		InstallationStateBucketRegion: w.cfg.InstallationStateBucketRegion,
		SandboxBucketName:             w.cfg.SandboxBucket,
		NuonAssumeRoleArn:             w.cfg.NuonAccessRoleArn,
	}

	_, err = execDestroyTerraform(ctx, act, dtReq)
	if err != nil {
		l.Debug("unable to execute terraform destroy: %w", err)
		err = fmt.Errorf("unable to run terraform destroy: %w", err)
		w.finishWithErr(ctx, req, act, "destroy_terraform", err)
		return resp, err
	}

	finishReq := FinishRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationStateBucket,
		Success:             true,
	}
	if _, err = execFinish(ctx, act, finishReq); err != nil {
		l.Debug("unable to execute finish step: %w", err)
		return resp, fmt.Errorf("unable to execute finish activity: %w", err)
	}

	l.Debug("finished deprovisioning installation", "response", resp)
	return resp, err
}

// execTerraformDestroy executes a terraform destroy activity
func execDestroyTerraform(ctx workflow.Context, act *Activities, req DestroyTerraformRequest) (DestroyTerraformResponse, error) {
	var resp DestroyTerraformResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing terraform destroy activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.DestroyTerraform, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// exec start executes the start activity
func execStart(ctx workflow.Context, act *Activities, req StartRequest) (StartResponse, error) {
	var resp StartResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing start", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.Start, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// exec finish executes the finish activity
func execFinish(ctx workflow.Context, act *Activities, req FinishRequest) (FinishResponse, error) {
	var resp FinishResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing finish", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishDeprovision, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
