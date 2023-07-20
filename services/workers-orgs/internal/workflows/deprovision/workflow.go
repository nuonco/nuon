package deprovision

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/kube"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/iam"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg internal.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg internal.Config
}

func (w *wkflow) Deprovision(ctx workflow.Context, req *orgsv1.DeprovisionRequest) (*orgsv1.DeprovisionResponse, error) {
	resp := &orgsv1.DeprovisionResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	act := NewActivities()

	iamReq := &iamv1.DeprovisionIAMRequest{
		OrgId: req.OrgId,
	}
	_, err := w.deprovisionIAM(ctx, iamReq)
	if err != nil {
		return resp, fmt.Errorf("unable to deprovision iam: %w", err)
	}

	_, err = uninstallWaypoint(ctx, act, UninstallWaypointRequest{
		Namespace:   req.OrgId,
		ReleaseName: fmt.Sprintf("wp-%s", req.OrgId),
		ClusterInfo: kube.ClusterInfo{
			ID:             w.cfg.OrgsK8sClusterID,
			Endpoint:       w.cfg.OrgsK8sPublicEndpoint,
			CAData:         w.cfg.OrgsK8sCAData,
			TrustedRoleARN: w.cfg.OrgsK8sRoleArn,
		},
	})
	if err != nil {
		return resp, fmt.Errorf("failed to uninstall waypoint: %w", err)
	}

	_, err = destroyNamespace(ctx, act, DestroyNamespaceRequest{NamespaceName: req.OrgId,
		ClusterInfo: kube.ClusterInfo{
			ID:             w.cfg.OrgsK8sClusterID,
			Endpoint:       w.cfg.OrgsK8sPublicEndpoint,
			CAData:         w.cfg.OrgsK8sCAData,
			TrustedRoleARN: w.cfg.OrgsK8sRoleArn,
		},
	})
	if err != nil {
		return resp, fmt.Errorf("failed to destroy namespace: %w", err)
	}

	l.Debug("finished teardown", "response", resp)
	return resp, nil
}

func (w *wkflow) deprovisionIAM(ctx workflow.Context, req *iamv1.DeprovisionIAMRequest) (*iamv1.DeprovisionIAMResponse, error) {
	var resp iamv1.DeprovisionIAMResponse

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Hour * 10,
		WorkflowTaskTimeout:      time.Hour * 1,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := iam.NewWorkflow(w.cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.DeprovisionIAM, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, err
	}

	return &resp, nil
}

func destroyNamespace(ctx workflow.Context, act *Activities, dnr DestroyNamespaceRequest) (DestroyNamespaceResponse, error) {
	var resp DestroyNamespaceResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing destroy namespace activity")
	fut := workflow.ExecuteActivity(ctx, act.DestroyNamespace, dnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func uninstallWaypoint(ctx workflow.Context, act *Activities, uwr UninstallWaypointRequest) (UninstallWaypointResponse, error) {
	var resp UninstallWaypointResponse
	if err := uwr.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	l := workflow.GetLogger(ctx)

	l.Debug("executing uninstall waypoint activity")
	fut := workflow.ExecuteActivity(ctx, act.UninstallWaypoint, uwr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
