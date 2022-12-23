package sandbox

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	sandboxv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/sandbox/v1"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

const (
	clusterIDKey       = "cluster_name"
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
		BackendBucketRegion:     w.cfg.InstallationsBucket,
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
