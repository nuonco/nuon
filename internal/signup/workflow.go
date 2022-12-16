package signup

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-waypoint"
	orgsv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1"
	runnerv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/runner/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/powertoolsdev/workers-orgs/internal/signup/runner"
	"github.com/powertoolsdev/workers-orgs/internal/signup/server"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) Signup(ctx workflow.Context, req *orgsv1.SignupRequest) (*orgsv1.SignupResponse, error) {
	resp := &orgsv1.SignupResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	id, err := shortid.ParseString(req.OrgId)
	if err != nil {
		return resp, fmt.Errorf("failed to generate short ID: %w", err)
	}
	waypointServerAddr := waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, id)

	act := NewActivities(nil)

	sendNotification(ctx, act, SendNotificationRequest{
		ID:      id,
		Started: true,
	})

	l.Debug("provisioning waypoint org server")
	_, err = provisionWaypointServer(ctx, w.cfg, server.ProvisionRequest{
		OrgID:  id,
		Region: req.Region,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to install runner: %w", err)
	}

	l.Debug("installing waypoint org runner")
	_, err = installWaypointRunner(ctx, w.cfg, &runnerv1.InstallRunnerRequest{
		OrgId: id,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to install runner: %w", err)
	}

	l.Debug("sending success notification")
	sendNotification(ctx, act, SendNotificationRequest{
		ID:                    id,
		Finished:              true,
		WaypointServerAddress: waypointServerAddr,
	})

	l.Debug("finished signup", "response", resp)
	return resp, nil
}

func provisionWaypointServer(
	ctx workflow.Context,
	cfg workers.Config,
	req server.ProvisionRequest,
) (server.ProvisionResponse, error) {
	var resp server.ProvisionResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing install waypoint workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := server.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.Provision, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func sendNotification(ctx workflow.Context, act *Activities, snr SendNotificationRequest) {
	var resp SendNotificationResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing send notification request activity")
	fut := workflow.ExecuteActivity(ctx, act.SendNotification, snr)

	if err := fut.Get(ctx, &resp); err != nil {
		l.Error("error sending notification: %s", err)
	}
}

func installWaypointRunner(
	ctx workflow.Context,
	cfg workers.Config,
	iwrr *runnerv1.InstallRunnerRequest,
) (*runnerv1.InstallRunnerResponse, error) {
	var resp runnerv1.InstallRunnerResponse

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := runner.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.Install, iwrr)

	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, err
	}

	return &resp, nil
}
