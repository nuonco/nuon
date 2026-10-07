package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/oci/signature"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateBuildSignatureVerificationRequest struct {
	BuildID      string `validate:"required"`
	JobID        string `validate:"required"`
	Required     bool
	JobSucceeded bool
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateBuildSignatureVerification(ctx context.Context, req *UpdateBuildSignatureVerificationRequest) error {
	var outcome app.ComponentBuildSignatureVerification
	switch {
	case !req.Required:
		outcome = app.ComponentBuildSignatureVerificationNotRequired
	case req.JobSucceeded:
		outcome = app.ComponentBuildSignatureVerificationVerified
	default:
		rejected, err := a.jobRejectedSignature(ctx, req.JobID)
		if err != nil {
			return err
		}
		if !rejected {
			return nil
		}
		outcome = app.ComponentBuildSignatureVerificationRejected
	}

	res := a.db.WithContext(ctx).
		Model(&app.ComponentBuild{ID: req.BuildID}).
		Updates(app.ComponentBuild{SignatureVerification: outcome})
	if res.Error != nil {
		return fmt.Errorf("unable to update build signature verification: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no build found: %s %w", req.BuildID, gorm.ErrRecordNotFound)
	}
	return nil
}

func (a *Activities) jobRejectedSignature(ctx context.Context, jobID string) (bool, error) {
	var execution app.RunnerJobExecution
	res := a.db.WithContext(ctx).
		Where(app.RunnerJobExecution{RunnerJobID: jobID}).
		Order("created_at DESC").
		First(&execution)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if res.Error != nil {
		return false, fmt.Errorf("unable to get runner job execution: %w", res.Error)
	}

	var result app.RunnerJobExecutionResult
	res = a.db.WithContext(ctx).
		Where(app.RunnerJobExecutionResult{RunnerJobExecutionID: execution.ID}).
		First(&result)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if res.Error != nil {
		return false, fmt.Errorf("unable to get runner job execution result: %w", res.Error)
	}

	return isSignatureRejection(result.ErrorMetadata), nil
}

func isSignatureRejection(meta pgtype.Hstore) bool {
	step, ok := meta["step"]
	return ok && step != nil && *step == signature.VerifyStep
}
