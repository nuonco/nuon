package deprovision

import "go.temporal.io/sdk/workflow"

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

// exec delete namespace will delete a namespace of choice for the install
func execDeleteNamespace(ctx workflow.Context, act *Activities, dnr DeleteNamespaceRequest) (DeleteNamespaceResponse, error) {
	var resp DeleteNamespaceResponse

	l := workflow.GetLogger(ctx)
	l.Debug("executing delete namespace activity")
	fut := workflow.ExecuteActivity(ctx, act.DeleteNamespace, dnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// exec list namespaces activity
func execListNamespaces(ctx workflow.Context, act *Activities, lnr ListNamespacesRequest) (ListNamespacesResponse, error) {
	var resp ListNamespacesResponse

	l := workflow.GetLogger(ctx)
	l.Debug("executing list namespaces activity")
	fut := workflow.ExecuteActivity(ctx, act.ListNamespaces, lnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
