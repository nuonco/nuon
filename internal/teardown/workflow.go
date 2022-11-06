package teardown

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/shortid"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type TeardownRequest struct {
	OrgID  string
	Region string
}
type TeardownResponse struct{}

func Teardown(ctx workflow.Context, req TeardownRequest) (TeardownResponse, error) {
	resp := TeardownResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	id, err := shortid.ParseString(req.OrgID)
	if err != nil {
		return resp, fmt.Errorf("failed to generate short ID: %w", err)
	}

	act := NewActivities()

	_, err = uninstallWaypoint(ctx, act, UninstallWaypointRequest{
		Namespace:   id,
		ReleaseName: fmt.Sprintf("wp-%s", id),
	})
	if err != nil {
		return resp, fmt.Errorf("failed to uninstall waypoint: %w", err)
	}

	_, err = destroyNamespace(ctx, act, DestroyNamespaceRequest{NamespaceName: id})
	if err != nil {
		return resp, fmt.Errorf("failed to destroy namespace: %w", err)
	}

	l.Debug("finished teardown", "response", resp)
	return resp, nil
}

func destroyNamespace(ctx workflow.Context, act *Activities, dnr DestroyNamespaceRequest) (DestroyNamespaceResponse, error) {
	var resp DestroyNamespaceResponse

	l := workflow.GetLogger(ctx)

	if err := validateDestroyNamespaceRequest(dnr); err != nil {
		return resp, err
	}
	l.Debug("executing destroy namespace activity")
	fut := workflow.ExecuteActivity(ctx, act.DestroyNamespace, dnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func uninstallWaypoint(ctx workflow.Context, act *Activities, uwr UninstallWaypointRequest) (UninstallWaypointResponse, error) {
	var resp UninstallWaypointResponse

	l := workflow.GetLogger(ctx)

	if err := validateUninstallWaypointRequest(uwr); err != nil {
		return resp, err
	}
	l.Debug("executing uninstall waypoint activity")
	fut := workflow.ExecuteActivity(ctx, act.UninstallWaypoint, uwr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
