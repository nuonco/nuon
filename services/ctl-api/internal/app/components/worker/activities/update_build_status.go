package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/redact"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/all"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type UpdateBuildStatus struct {
	BuildID           string                   `validate:"required"`
	Status            app.ComponentBuildStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
	Error             string                   `json:",omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateBuildStatus(ctx context.Context, req UpdateBuildStatus) error {
	currentApp := app.ComponentBuild{
		ID: req.BuildID,
	}
	var failure *compositeerrors.CompositeErrorData
	res := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.Error != "" {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id", "status", "status_description", "composite_error").
				Where(&currentApp).Take(&currentApp).Error; err != nil {
				return fmt.Errorf("unable to get failed build: %w", err)
			}
			switch currentApp.Status {
			case app.ComponentBuildStatusPolicyFailed, app.ComponentBuildStatus(app.StatusCancelled):
				req.Status = currentApp.Status
				req.StatusDescription = currentApp.StatusDescription
			}
			failure = currentApp.CompositeError
			if failure == nil {
				var err error
				failure, err = errparse.ParseRunnerJobResult(false, map[string]string{errparse.ErrorMetadataMessage: req.Error}, &app.RunnerJob{
					OwnerType: "component_builds", OwnerID: req.BuildID,
				}, nil)
				if err != nil {
					return fmt.Errorf("unable to record build failure: %w", err)
				}
				if failure != nil {
					failure.Message = redact.DiagnosticSecrets(req.StatusDescription)
				}
			}
		}
		res := tx.Model(&currentApp).Updates(app.ComponentBuild{
			Status:            req.Status,
			StatusDescription: redact.DiagnosticSecrets(req.StatusDescription),
			CompositeError:    failure,
		})
		if res.Error != nil {
			return fmt.Errorf("unable to update build: %w", res.Error)
		}
		if res.RowsAffected < 1 {
			return fmt.Errorf("no build found: %s %w", req.BuildID, gorm.ErrRecordNotFound)
		}
		return nil
	})
	return res
}

func FailBuild(ctx workflow.Context, buildID string, buildErr error) error {
	if buildErr == nil || temporal.IsCanceledError(buildErr) {
		return buildErr
	}
	ctx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	status := app.ComponentBuildStatusError
	description := "Build failed: " + compositeerrors.ErrorMessage(buildErr)
	var applicationErr *temporal.ApplicationError
	if errors.As(buildErr, &applicationErr) && applicationErr.Type() == "component-build-policy" {
		status = app.ComponentBuildStatusPolicyFailed
		description = compositeerrors.ErrorMessage(buildErr)
	}
	description = redact.DiagnosticSecrets(description)
	description = compositeerrors.TruncateRunes(description, 499)
	statusErr := AwaitUpdateBuildStatus(ctx, UpdateBuildStatus{
		BuildID: buildID, Status: status,
		StatusDescription: description, Error: redact.DiagnosticSecrets(buildErr.Error()),
	})
	if statusErr != nil {
		statusErr = fmt.Errorf("unable to persist build failure: %w", statusErr)
	}
	historyErr := statusactivities.AwaitUpdateBuildStatusV2(ctx, statusactivities.UpdateBuildStatusV2{
		BuildID: buildID, Status: status, StatusDescription: description,
	})
	if historyErr != nil {
		historyErr = fmt.Errorf("unable to persist build failure history: %w", historyErr)
	}
	if statusErr != nil || historyErr != nil {
		return errors.Join(buildErr, statusErr, historyErr)
	}
	return buildErr
}
