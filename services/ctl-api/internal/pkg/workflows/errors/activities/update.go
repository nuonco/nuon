package erroractivities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/structured_errors"
)

type AppendErrorsRequest struct {
	ID     string                            `json:"id" validate:"required"`
	Errors structured_errors.CompositeErrors `json:"errors" validate:"required"`
}

type ClearErrorsRequest struct {
	ID string `json:"id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) AppendDeployErrors(ctx context.Context, req AppendErrorsRequest) error {
	return a.appendErrors(ctx, &app.InstallDeploy{}, req.ID, req.Errors)
}

// @temporal-gen-v2 activity
func (a *Activities) ClearDeployErrors(ctx context.Context, req ClearErrorsRequest) error {
	return a.clearErrors(ctx, &app.InstallDeploy{}, req.ID)
}

// @temporal-gen-v2 activity
func (a *Activities) AppendBuildErrors(ctx context.Context, req AppendErrorsRequest) error {
	return a.appendErrors(ctx, &app.ComponentBuild{}, req.ID, req.Errors)
}

// @temporal-gen-v2 activity
func (a *Activities) ClearBuildErrors(ctx context.Context, req ClearErrorsRequest) error {
	return a.clearErrors(ctx, &app.ComponentBuild{}, req.ID)
}

// @temporal-gen-v2 activity
func (a *Activities) AppendSandboxRunErrors(ctx context.Context, req AppendErrorsRequest) error {
	return a.appendErrors(ctx, &app.InstallSandboxRun{}, req.ID, req.Errors)
}

// @temporal-gen-v2 activity
func (a *Activities) ClearSandboxRunErrors(ctx context.Context, req ClearErrorsRequest) error {
	return a.clearErrors(ctx, &app.InstallSandboxRun{}, req.ID)
}

// @temporal-gen-v2 activity
func (a *Activities) AppendActionRunErrors(ctx context.Context, req AppendErrorsRequest) error {
	return a.appendErrors(ctx, &app.InstallActionWorkflowRun{}, req.ID, req.Errors)
}

// @temporal-gen-v2 activity
func (a *Activities) ClearActionRunErrors(ctx context.Context, req ClearErrorsRequest) error {
	return a.clearErrors(ctx, &app.InstallActionWorkflowRun{}, req.ID)
}

func (a *Activities) appendErrors(ctx context.Context, model any, id string, errs structured_errors.CompositeErrors) error {
	if err := structured_errors.Append(a.db.WithContext(ctx), model, id, errs); err != nil {
		return fmt.Errorf("unable to append errors to %T(%s): %w", model, id, err)
	}
	return nil
}

func (a *Activities) clearErrors(ctx context.Context, model any, id string) error {
	if err := structured_errors.Clear(a.db.WithContext(ctx), model, id); err != nil {
		return fmt.Errorf("unable to clear errors on %T(%s): %w", model, id, err)
	}
	return nil
}
