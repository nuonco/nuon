package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateInstallStackVersionRunRequest struct {
	// Status must be "succeeded" or "failed"; runs are created as "running".
	Status            app.Status     `json:"status" validate:"required"`
	StatusDescription string         `json:"status_description"`
	Data              map[string]any `json:"data"`
}

// @ID						UpdateInstallStackVersionRun
// @Summary				update a stack version run
// @Description			mark a run terminal (succeeded/failed). Public endpoint, mirrors phone-home: phone_home_id in the URL acts as the secret.
// @Param					install_id		path	string	true	"install ID"
// @Param					phone_home_id	path	string	true	"stack version phone-home ID (used as the URL secret)"
// @Param					run_id			path	string	true	"run ID"
// @Param					req				body	UpdateInstallStackVersionRunRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallStackVersionRun
// @Router					/v1/installs/{install_id}/stack-runs/{phone_home_id}/{run_id} [patch]
func (s *service) UpdateInstallStackVersionRun(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	phoneHomeID := ctx.Param("phone_home_id")
	runID := ctx.Param("run_id")

	var req UpdateInstallStackVersionRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if req.Status != app.InstallStackVersionRunStatusSucceeded && req.Status != app.InstallStackVersionRunStatusFailed {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid status %q", req.Status),
			Description: "status must be 'succeeded' or 'failed'",
		})
		return
	}

	// Validate (install_id, phone_home_id) → stack version, then run.
	var stackVersion app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{InstallID: installID, PhoneHomeID: phoneHomeID}).
		First(&stackVersion); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "stack version not found"})
			return
		}
		ctx.Error(fmt.Errorf("load stack version: %w", res.Error))
		return
	}

	var run app.InstallStackVersionRun
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersionRun{ID: runID, InstallStackVersionID: stackVersion.ID}).
		First(&run); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "run not found for stack version"})
			return
		}
		ctx.Error(fmt.Errorf("load run: %w", res.Error))
		return
	}

	if run.Status.Status == req.Status {
		ctx.JSON(http.StatusOK, run)
		return
	}
	if run.Status.Status != app.InstallStackVersionRunStatusRunning {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("cannot transition from %q to %q", run.Status.Status, req.Status),
			Description: "run is already terminal",
		})
		return
	}

	updates := app.InstallStackVersionRun{
		Status: app.NewCompositeStatus(ctx, req.Status),
	}
	if len(req.Data) > 0 {
		updates.Data = generics.ToHstore(pkggenerics.ToStringMap(req.Data))
	}
	if res := s.db.WithContext(ctx).Model(&app.InstallStackVersionRun{ID: runID}).Updates(updates); res.Error != nil {
		ctx.Error(fmt.Errorf("update run: %w", res.Error))
		return
	}

	parentStatus := app.InstallStackVersionStatusActive
	if req.Status == app.InstallStackVersionRunStatusFailed {
		parentStatus = app.StatusError
	}
	if res := s.db.WithContext(ctx).
		Model(&app.InstallStackVersion{ID: stackVersion.ID}).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(ctx, parentStatus),
		}); res.Error != nil {
		ctx.Error(fmt.Errorf("cascade stack version status: %w", res.Error))
		return
	}

	// Close the run's log stream so the dashboard's SSE tailer stops polling
	// and switches to paginated GET for replay.
	if run.LogStreamID != "" {
		if res := s.db.WithContext(ctx).
			Model(&app.LogStream{ID: run.LogStreamID}).
			Update("open", false); res.Error != nil {
			ctx.Error(fmt.Errorf("close run log stream: %w", res.Error))
			return
		}
	}

	if res := s.db.WithContext(ctx).First(&run, "id = ?", runID); res.Error != nil {
		ctx.Error(fmt.Errorf("reload run: %w", res.Error))
		return
	}
	ctx.JSON(http.StatusOK, run)
}
