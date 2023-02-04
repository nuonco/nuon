package sandbox

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/mitchellh/mapstructure"
	sandboxv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/sandbox/v1"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// ProvisionSandbox is a workflow that creates an app install sandbox using terraform
func (w wkflow) ProvisionSandbox(ctx workflow.Context, req *sandboxv1.ProvisionSandboxRequest) (*sandboxv1.ProvisionSandboxResponse, error) {
	resp := &sandboxv1.ProvisionSandboxResponse{}
	l := workflow.GetLogger(ctx)

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// NOTE(jdt): this is just so that we can use the method names
	// the actual struct isn't used by temporal during dispatch at all
	act := NewActivities(workers.Config{})

	psReq := ApplySandboxRequest{
		AppID:                   req.AppId,
		OrgID:                   req.OrgId,
		InstallID:               req.InstallId,
		BackendBucketName:       w.cfg.InstallationsBucket,
		BackendBucketRegion:     w.cfg.InstallationsBucketRegion,
		AccountSettings:         req.AccountSettings,
		SandboxSettings:         req.SandboxSettings,
		SandboxBucketName:       w.cfg.SandboxBucket,
		NuonAccessRoleArn:       w.cfg.NuonAccessRoleArn,
		OrgInstanceRoleTemplate: w.cfg.OrgInstanceRoleTemplate,
	}
	psr, err := provisionSandbox(ctx, act, psReq)
	if err != nil {
		err = fmt.Errorf("unable to provision sandbox: %w", err)
		return resp, err
	}
	tfOutputs, err := ParseTerraformOutputs(psr.Outputs)
	if err != nil {
		err = fmt.Errorf("unable to parse terraform outputs: %w", err)
		return resp, err
	}
	var respTfOutputs map[string]string
	if err := mapstructure.Decode(tfOutputs, &respTfOutputs); err != nil {
		err = fmt.Errorf("unable to parse tf outputs to proto output: %w", err)
		return resp, err
	}
	resp.TerraformOutputs = respTfOutputs

	l.Debug("finished provisioning", "response", resp)
	return resp, nil
}

// provisionSandbox executes a provision sandbox activity
func provisionSandbox(ctx workflow.Context, act *Activities, req ApplySandboxRequest) (ApplySandboxResponse, error) {
	var resp ApplySandboxResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing provision sandbox activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ApplySandbox, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
