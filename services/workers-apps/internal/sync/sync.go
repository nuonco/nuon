package sync

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	appv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
)

type Workflow struct {
	cfg  workers.Config
	v    *validator.Validate
	acts *Activities
}

func NewWorkflow(v *validator.Validate, cfg workers.Config) Workflow {
	return Workflow{
		cfg:  cfg,
		v:    v,
		acts: &Activities{},
	}
}

func (w Workflow) Sync(ctx workflow.Context, req *appv1.SyncRequest) (*appv1.SyncResponse, error) {
	resp := appv1.SyncResponse{}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	// run terraform
	if err := w.execTerraformRun(ctx, &ExecTerraformRequest{
		TerraformJSON:    req.TerraformJson,
		TerraformVersion: req.TerraformVersion,
		APIURL:           req.ApiUrl,
		APIToken:         req.ApiToken,
		OrgID:            req.OrgId,
		AppID:            req.AppId,

		// Backend
		BackendIAMRoleARN: fmt.Sprintf(w.cfg.OrgsRoleTemplate, req.OrgId),
		BackendKey:        prefix.AppConfigPath(req.OrgId, req.AppId),
		BackendRegion:     "us-west-2",
		BackendBucket:     w.cfg.OrgsBucketName,
	}); err != nil {
		return nil, fmt.Errorf("unable to execute terraform: %w", err)
	}

	return &resp, nil
}

func (w *Workflow) execTerraformRun(
	ctx workflow.Context,
	req *ExecTerraformRequest,
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, w.acts.ExecTerraform, req)
	var respErr error
	if err := fut.Get(ctx, &respErr); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}

	if respErr != nil {
		return fmt.Errorf("activity returned error: %w", respErr)
	}

	return nil
}
