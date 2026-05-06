package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// LogStreamOwnerTypeInstallStackVersionRuns identifies log streams created
// for SDK provisioner runs. Used as the OwnerType when the dashboard queries
// for the stream's logs.
const LogStreamOwnerTypeInstallStackVersionRuns = "install_stack_version_runs"

// @ID						CreateInstallStackVersionRun
// @Summary				create a stack version run
// @Description			start a new run for an install stack version. Public endpoint, mirrors phone-home: the per-stack-version phone_home_id in the URL acts as the secret. Used by the AWS-native SDK provisioner; legacy CFN/TF flows use phone-home.
// @Param					install_id		path	string	true	"install ID"
// @Param					phone_home_id	path	string	true	"stack version phone-home ID (used as the URL secret)"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.InstallStackVersionRun
// @Router					/v1/installs/{install_id}/stack-runs/{phone_home_id} [post]
func (s *service) CreateInstallStackVersionRun(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	phoneHomeID := ctx.Param("phone_home_id")

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

	// Public endpoint: org context isn't set by middleware, but the helper
	// chain (LogStream.BeforeCreate, AddAccountOrgRole, etc.) needs it. Set
	// it from the stack version we just loaded.
	reqCtx := cctx.SetOrgIDContext(ctx.Request.Context(), stackVersion.OrgID)
	reqCtx = cctx.SetAccountIDContext(reqCtx, stackVersion.CreatedByID)

	run := app.InstallStackVersionRun{
		InstallStackVersionID: stackVersion.ID,
		OrgID:                 stackVersion.OrgID,
		Status:                app.NewCompositeStatus(reqCtx, app.InstallStackVersionRunStatusRunning),
	}
	if res := s.db.WithContext(reqCtx).Create(&run); res.Error != nil {
		ctx.Error(fmt.Errorf("create stack version run: %w", res.Error))
		return
	}

	logStream, err := s.helpers.CreateLogStream(reqCtx, LogStreamOwnerTypeInstallStackVersionRuns, run.ID, "")
	if err != nil {
		ctx.Error(fmt.Errorf("create run log stream: %w", err))
		return
	}
	if res := s.db.WithContext(reqCtx).
		Model(&app.InstallStackVersionRun{ID: run.ID}).
		Update("log_stream_id", logStream.ID); res.Error != nil {
		ctx.Error(fmt.Errorf("attach log stream to run: %w", res.Error))
		return
	}
	run.LogStreamID = logStream.ID
	run.LogStream = logStream

	if res := s.db.WithContext(reqCtx).
		Model(&app.InstallStackVersion{ID: stackVersion.ID}).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(reqCtx, app.InstallStackVersionStatusProvisioning),
		}); res.Error != nil {
		ctx.Error(fmt.Errorf("update stack version status: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, run)
}
