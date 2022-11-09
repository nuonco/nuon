package sandbox

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	workers "github.com/powertoolsdev/workers-installs/internal"
)

// ProvisionRequest includes the set of arguments needed to provision a sandbox
type ProvisionRequest struct {
	OrgID     string `json:"org_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	InstallID string `json:"install_id" validate:"required"`

	AccountSettings *AccountSettings `json:"account_settings" validate:"required"`

	SandboxSettings struct {
		Name    string `json:"name" validate:"required"`
		Version string `json:"version" validate:"required"`
	} `json:"sandbox_settings" validate:"required"`
}

func (p ProvisionRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type ProvisionResponse struct {
	TerraformOutputs map[string]string
}

const (
	clusterIDKey       = "cluster_id"
	clusterEndpointKey = "cluster_endpoint"
	clusterCAKey       = "cluster_certificate_authority_data"
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
func (w wkflow) ProvisionSandbox(ctx workflow.Context, req ProvisionRequest) (ProvisionResponse, error) {
	resp := ProvisionResponse{}
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
		AppID:               req.AppID,
		OrgID:               req.OrgID,
		InstallID:           req.InstallID,
		BackendBucketName:   w.cfg.InstallationStateBucket,
		BackendBucketRegion: w.cfg.InstallationStateBucketRegion,
		AccountSettings:     req.AccountSettings,
		SandboxSettings:     req.SandboxSettings,
		SandboxBucketName:   w.cfg.SandboxBucket,
		NuonAccessRoleArn:   w.cfg.NuonAccessRoleArn,
	}
	psr, err := provisionSandbox(ctx, act, psReq)
	if err != nil {
		err = fmt.Errorf("unable to provision sandbox: %w", err)
		return resp, err
	}
	resp.TerraformOutputs = psr.Outputs

	if err = checkKeys(psr.Outputs, []string{clusterIDKey, clusterEndpointKey, clusterCAKey}); err != nil {
		err = fmt.Errorf("missing necessary TF output to continue: %w", err)
		return resp, err
	}

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

func checkKeys(m map[string]string, keys []string) error {
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return fmt.Errorf("missing key: %s", k)
		}
	}
	return nil
}
