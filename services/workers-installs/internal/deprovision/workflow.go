package deprovision

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/generics"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
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
	act := NewActivities(nil, nil, nil)

	stReq := StartRequest{
		DeprovisionRequest:  req,
		InstallationsBucket: w.cfg.InstallationsBucket,
	}
	_, err := execStart(ctx, act, stReq)
	if err != nil {
		l.Debug("unable to execute start activity", "error", err)
		return resp, fmt.Errorf("unable to execute start activity: %w", err)
	}

	// NOTE(jm): this is not a long term solution, eventually we will manage both the runner and the different
	// components using nuon components, and then will just remove these by orchestrating the executors upstream.
	//
	// however, for now, until this all works we just "cheat" and delete the builtin namespace
	listResp, err := execListNamespaces(ctx, act, ListNamespacesRequest{
		AppID:     req.AppId,
		OrgID:     req.OrgId,
		InstallID: req.InstallId,
	})
	if err != nil {
		l.Debug("unable to list namespaces", "error", err)
		return resp, fmt.Errorf("unable to delete namespace: %w", err)
	}

	for _, namespace := range listResp.Namespaces {
		if generics.SliceContains(namespace, terraformManagedNamespaces) {
			continue
		}

		_, err = execDeleteNamespace(ctx, act, DeleteNamespaceRequest{
			AppID:     req.AppId,
			OrgID:     req.OrgId,
			InstallID: req.InstallId,
			Namespace: namespace,
		})
		if err != nil {
			l.Debug("unable to delete namespace activity", "error", err)
			return resp, fmt.Errorf("unable to delete namespace: %w", err)
		}
	}
	if req.PlanOnly {
		l.Info("skipping the rest of the workflow - plan only")
		return resp, nil
	}

	if err := w.deprovisionNoopBuild(ctx, req); err != nil {
		err = fmt.Errorf("unable to create noop build: %w", err)
		return resp, err
	}

	_, err = w.deprovisionSandbox(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to deprovision sandbox: %w", err)
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
