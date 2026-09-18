package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type UpdateAppBranchRunBuildsCompletedInput struct {
	RunID           string `json:"run_id" validate:"required"`
	BuildsCompleted bool   `json:"builds_completed"`
}

type UpdateAppBranchRunBuildsCompletedOutput struct {
	CompositeError *compositeerrors.CompositeErrorData `json:"composite_error,omitempty" temporaljson:"composite_error,omitzero,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateAppBranchRunBuildsCompleted(ctx context.Context, input *UpdateAppBranchRunBuildsCompletedInput) (*UpdateAppBranchRunBuildsCompletedOutput, error) {
	if err := a.v.Struct(input); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	var run app.AppBranchRun
	if err := a.db.WithContext(ctx).First(&run, "id = ?", input.RunID).Error; err != nil {
		return nil, errors.Wrap(err, "app branch run not found")
	}

	value := "false"
	if input.BuildsCompleted {
		value = "true"
	}

	if run.Labels == nil {
		run.Labels = labels.Labels{}
	}
	run.Labels.Merge(labels.Labels{
		app.AppBranchRunLabelBuildsCompleted: value,
	})

	if err := a.db.WithContext(ctx).
		Model(&run).
		Select("labels").
		Updates(&run).Error; err != nil {
		return nil, fmt.Errorf("unable to update builds_completed label on run %s: %w", input.RunID, err)
	}

	out := &UpdateAppBranchRunBuildsCompletedOutput{}
	if input.BuildsCompleted {
		return out, nil
	}

	data, err := a.resolveBuildsCompositeError(ctx, input.RunID)
	if err != nil {
		a.l.Warn("unable to resolve app branch run builds composite error",
			zap.String("app_branch_run_id", input.RunID),
			zap.Error(err))
		return out, nil
	}
	if data == nil {
		return out, nil
	}
	if err := a.setAppBranchRunCompositeError(ctx, input.RunID, data); err != nil {
		return nil, err
	}
	out.CompositeError = data
	return out, nil
}

func (a *Activities) resolveBuildsCompositeError(ctx context.Context, runID string) (*compositeerrors.CompositeErrorData, error) {
	var componentBuilds []app.ComponentBuild
	if err := a.db.WithContext(ctx).
		Select("id", "status", "composite_error", "created_at").
		Where("app_branch_run_id = ?", runID).
		Where("status IN ?", []app.ComponentBuildStatus{
			app.ComponentBuildStatusError,
			app.ComponentBuildStatusPolicyFailed,
		}).
		Order("created_at asc").
		Find(&componentBuilds).Error; err != nil {
		return nil, fmt.Errorf("unable to get failed component builds for run %s: %w", runID, err)
	}
	for _, build := range componentBuilds {
		data, err := a.buildCompositeError(ctx, build.CompositeError, build.ID, "component_builds")
		if err != nil {
			return nil, err
		}
		if data != nil {
			return data, nil
		}
	}

	var sandboxBuilds []app.AppSandboxBuild
	if err := a.db.WithContext(ctx).
		Select("id", "status", "composite_error", "created_at").
		Where("app_branch_run_id = ?", runID).
		Where(&app.AppSandboxBuild{Status: app.AppSandboxBuildStatusError}).
		Order("created_at asc").
		Find(&sandboxBuilds).Error; err != nil {
		return nil, fmt.Errorf("unable to get failed sandbox builds for run %s: %w", runID, err)
	}
	for _, build := range sandboxBuilds {
		data, err := a.buildCompositeError(ctx, build.CompositeError, build.ID, "app_sandbox_builds")
		if err != nil {
			return nil, err
		}
		if data != nil {
			return data, nil
		}
	}

	return nil, nil
}

func (a *Activities) buildCompositeError(ctx context.Context, rowError *compositeerrors.CompositeErrorData, buildID, ownerType string) (*compositeerrors.CompositeErrorData, error) {
	if rowError != nil && rowError.Type != "" {
		return rowError, nil
	}

	jobError, err := runnershelpers.GetLatestJobCompositeError(ctx, a.db, runnershelpers.GetLatestJobCompositeErrorRequest{
		OwnerID:   buildID,
		OwnerType: ownerType,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get composite error for %s %s: %w", ownerType, buildID, err)
	}
	return jobError, nil
}
