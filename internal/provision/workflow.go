package provision

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-waypoint"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
	defaultDeployTimeout   = time.Minute * 15
)

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

func (w *Workflow) Provision(ctx workflow.Context, req *instancesv1.ProvisionRequest) (*instancesv1.ProvisionResponse, error) {
	resp := &instancesv1.ProvisionResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(nil)

	shnReq := SendHostnameNotificationRequest{
		OrgID:                "todo-org-id",
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, "todo-org-id"),
		InstallID:            "todo-install-id",
		AppID:                "todo-app-id",
	}
	shnResp, err := execSendHostnameNotification(ctx, act, shnReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully sent hostname notification: ", shnResp)

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
