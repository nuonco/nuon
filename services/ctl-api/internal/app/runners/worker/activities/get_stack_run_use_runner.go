package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetStackRunUseRunnerRequest struct {
	InstallStackVersionRunID string `validate:"required"`
}

// GetStackRunUseRunner reports whether a stack run's outputs used a runner. It
// reads the raw output key rather than the per-cloud structs so it stays
// correct for any stack that grows the variable; a missing key means the
// stack predates use_runner and a runner is used.
//
// @temporal-gen-v2 activity
// @max-retries 1
// @by-field InstallStackVersionRunID
// @local
func (a *Activities) GetStackRunUseRunner(ctx context.Context, req GetStackRunUseRunnerRequest) (bool, error) {
	run := app.InstallStackVersionRun{}
	res := a.db.WithContext(ctx).
		Select("id", "data").
		First(&run, "id = ?", req.InstallStackVersionRunID)
	if res.Error != nil {
		return false, dbgenerics.TemporalGormError(res.Error, "unable to get install stack version run")
	}

	val, ok := run.Data["use_runner"]
	if !ok || val == nil {
		return true, nil
	}

	return *val != "false", nil
}
