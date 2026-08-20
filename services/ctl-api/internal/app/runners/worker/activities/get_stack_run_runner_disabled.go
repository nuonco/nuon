package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetStackRunRunnerDisabledRequest struct {
	InstallStackVersionRunID string `validate:"required"`
}

// GetStackRunRunnerDisabled reports whether a stack run's outputs disabled the
// runner. It reads the raw output key rather than the per-cloud structs so it
// stays correct for any stack that grows the variable; a missing key means the
// stack predates runner_enabled and the runner is enabled.
//
// @temporal-gen-v2 activity
// @max-retries 1
// @by-field InstallStackVersionRunID
// @local
func (a *Activities) GetStackRunRunnerDisabled(ctx context.Context, req GetStackRunRunnerDisabledRequest) (bool, error) {
	run := app.InstallStackVersionRun{}
	res := a.db.WithContext(ctx).
		Select("id", "data").
		First(&run, "id = ?", req.InstallStackVersionRunID)
	if res.Error != nil {
		return false, dbgenerics.TemporalGormError(res.Error, "unable to get install stack version run")
	}

	val, ok := run.Data["runner_enabled"]
	if !ok || val == nil {
		return false, nil
	}

	return *val == "false", nil
}
