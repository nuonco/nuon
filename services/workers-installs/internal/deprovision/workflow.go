package deprovision

import (
	"fmt"
	"time"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg:        cfg,
		sharedActs: activities.NewActivities(nil, nil),
		acts:       NewActivities(nil, nil),
	}
}

type wkflow struct {
	cfg        workers.Config
	sharedActs *activities.Activities
	acts       *Activities
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
	if req.PlanOnly {
		l.Info("skipping the rest of the workflow - plan only")
		return resp, nil
	}

	if err := w.deprovisionRunner(ctx, req); err != nil {
		err = fmt.Errorf("unable to deprovision runner: %w", err)
		l.Info("error deprovisioning runner", zap.Error(err))
		return resp, nil
	}

	if err := w.deprovisionNoopBuild(ctx, req); err != nil {
		err = fmt.Errorf("unable to create noop build: %w", err)
		return resp, err
	}

	_, err := w.deprovisionSandbox(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to deprovision sandbox: %w", err)
		return resp, err
	}

	l.Debug("finished deprovisioning installation", "response", resp)
	return resp, err
}
