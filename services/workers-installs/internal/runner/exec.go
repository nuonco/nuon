package runner

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

// execute an aws activity
func (w *wkflow) execAWSActivity(ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	if err := w.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, resp); err != nil {
		return err
	}

	return nil
}

func (w *wkflow) execWaypointActivity(ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	if err := w.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, resp); err != nil {
		return err
	}

	return nil
}

// exec delete namespace will delete a namespace of choice for the install
func (w *wkflow) execDeleteNamespace(ctx workflow.Context, dnr DeleteNamespaceRequest) (DeleteNamespaceResponse, error) {
	var resp DeleteNamespaceResponse

	l := workflow.GetLogger(ctx)
	l.Debug("executing delete namespace activity")
	fut := workflow.ExecuteActivity(ctx, w.act.DeleteNamespace, dnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// exec list namespaces activity
func (w *wkflow) execListNamespaces(ctx workflow.Context, lnr ListNamespacesRequest) (ListNamespacesResponse, error) {
	var resp ListNamespacesResponse

	l := workflow.GetLogger(ctx)
	l.Debug("executing list namespaces activity")
	fut := workflow.ExecuteActivity(ctx, w.act.ListNamespaces, lnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointProject executes an activity to create the waypoint project on the org's server
func (w *wkflow) createWaypointProject(ctx workflow.Context, req CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint project activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateWaypointProject, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// getWaypointServerCookie executes an activity to get the the waypoint server
func (w *wkflow) getWaypointServerCookie(ctx workflow.Context, req GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
	var resp GetWaypointServerCookieResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing get waypoint server cookie")
	fut := workflow.ExecuteActivity(ctx, w.act.GetWaypointServerCookie, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// installWaypoint executes an activity to install waypoint into the sandbox
func (w *wkflow) installWaypoint(ctx workflow.Context, req InstallWaypointRequest) (InstallWaypointResponse, error) {
	var resp InstallWaypointResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing install waypoint activity")
	fut := workflow.ExecuteActivity(ctx, w.act.InstallWaypoint, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// adoptWaypointRunner adopts the waypoint runner
func (w *wkflow) adoptWaypointRunner(ctx workflow.Context, req AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
	var resp AdoptWaypointRunnerResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing adopt waypoint runner activity")
	fut := workflow.ExecuteActivity(ctx, w.act.AdoptWaypointRunner, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointWorkspace creates a waypoint workspace
func (w *wkflow) createWaypointWorkspace(ctx workflow.Context, req CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
	var resp CreateWaypointWorkspaceResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint workspace activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateWaypointWorkspace, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createRoleBinding: creates the rolebinding in the correct namespace
func (w *wkflow) createRoleBinding(
	ctx workflow.Context,
	req CreateRoleBindingRequest,
) (CreateRoleBindingResponse, error) {
	var resp CreateRoleBindingResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create role binding activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateRoleBinding, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointRunnerProfile: creates the runner profile for this install
func (w *wkflow) createWaypointRunnerProfile(
	ctx workflow.Context,
	req CreateWaypointRunnerProfileRequest,
) (CreateWaypointRunnerProfileResponse, error) {
	var resp CreateWaypointRunnerProfileResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint runner profile activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateWaypointRunnerProfile, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
