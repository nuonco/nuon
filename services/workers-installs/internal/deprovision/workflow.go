package deprovision

import (
	"fmt"
	"time"

	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/sandbox"
	"go.temporal.io/sdk/workflow"
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

func (w wkflow) finishWithErr(ctx workflow.Context, req *installsv1.DeprovisionRequest, act *Activities, step string, err error) {
	l := workflow.GetLogger(ctx)
	finishReq := FinishRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationsBucket,
		Success:             false,
		ErrorStep:           step,
		ErrorMessage:        fmt.Sprintf("%s", err),
	}

	if resp, execErr := execFinish(ctx, act, finishReq); execErr != nil {
		l.Debug("unable to finish with error", "error", execErr, "response", resp)
	}
}

// Deprovision method destroys the infrastructure for an installation
func (w wkflow) Deprovision(ctx workflow.Context, req *installsv1.DeprovisionRequest) (*installsv1.DeprovisionResponse, error) {
	resp := &installsv1.DeprovisionResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("validating deprovision request")
	if err := req.Validate(); err != nil {
		l.Debug("unable to validate terraform destroy request", "error", err)
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)
	act := NewActivities(nil)

	stReq := StartRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationsBucket,
	}
	_, err := execStart(ctx, act, stReq)
	if err != nil {
		l.Debug("unable to execute start activity", "error", err)
		return resp, fmt.Errorf("unable to execute start activity: %w", err)
	}

	cpReq := planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.SandboxInput{
				OrgId:     req.OrgId,
				AppId:     req.AppId,
				InstallId: req.InstallId,
				SandboxSettings: &planv1.SandboxSettings{
					Name:    req.SandboxSettings.Name,
					Version: req.SandboxSettings.Version,
				},
				TerraformVersion: req.SandboxSettings.TerraformVersion,
				Type:             planv1.SandboxInputType_SANDBOX_INPUT_TYPE_DEPROVISION,
				AccountSettings: &planv1.SandboxInput_Aws{
					Aws: &planv1.AWSSettings{
						Region:    req.AccountSettings.Region,
						AccountId: req.AccountSettings.AwsAccountId,
						RoleArn:   req.AccountSettings.AwsRoleArn,
					},
				},
				RootDomain: fmt.Sprintf("%s.%s", req.InstallId, w.cfg.PublicDomain),
			},
		},
	}

	l.Debug("executing sandbox plan")
	spr, err := sandbox.Plan(ctx, &cpReq)
	if err != nil {
		err = fmt.Errorf("unable to plan deprovision sandbox: %w", err)
		w.finishWithErr(ctx, req, act, "sandbox_plan", err)
		return resp, err
	}

	// if req.PlanOnly {
	//	l.Info("skipping the rest of the workflow - plan only")
	//	w.finishWorkflow(ctx, req, resp, nil)
	//	return resp, nil
	// }

	l.Debug("executing sandbox execute")
	seReq := executev1.ExecutePlanRequest{Plan: spr.Plan}
	_, err = sandbox.Execute(ctx, &seReq)
	if err != nil {
		err = fmt.Errorf("unable to execute deprovision sandbox: %w", err)
		w.finishWithErr(ctx, req, act, "sandbox_execute", err)
		return resp, err
	}

	finishReq := FinishRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationsBucket,
		Success:             true,
	}
	if _, err = execFinish(ctx, act, finishReq); err != nil {
		l.Debug("unable to execute finish step", "error", err)
		return resp, fmt.Errorf("unable to execute finish activity: %w", err)
	}

	l.Debug("finished deprovisioning installation", "response", resp)
	return resp, err
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
