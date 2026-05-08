package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	updateinstallstackoutputs "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/updateinstallstackoutputs"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
// @Router					/v1/stack-runs/{phone_home_id}/{run_id} [patch]
func (s *service) UpdateInstallStackVersionRun(ctx *gin.Context) {
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

	// phone_home_id is unique per stack version; install_id derives from the row.
	var stackVersion app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{PhoneHomeID: phoneHomeID}).
		First(&stackVersion); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "stack version not found"})
			return
		}
		ctx.Error(fmt.Errorf("load stack version: %w", res.Error))
		return
	}
	installID := stackVersion.InstallID

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

	// Parent status reflects what the run actually did. A failed run always
	// surfaces as StatusError. A succeeded deprovision flips the version to
	// Destroyed; provision/reprovision resolve to Active.
	var parentStatus app.Status
	switch {
	case req.Status == app.InstallStackVersionRunStatusFailed:
		parentStatus = app.StatusError
	case run.Kind == app.InstallStackVersionRunKindDeprovision:
		parentStatus = app.InstallStackVersionStatusDestroyed
	default:
		parentStatus = app.InstallStackVersionStatusActive
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

	// Trigger the same UpdateInstallStackOutputs flow the legacy phone-home
	// handler runs (post_install_phone_home.go:163). Without this, the
	// install_stack_outputs row stays empty and runner-auth/aws-iid 5xxs on
	// every IID auth request because it can't validate the runner's account
	// against the install. Provision and reprovision both populate outputs;
	// failed runs and deprovisions skip — failure shouldn't rewrite outputs,
	// and a deprovision should leave them intact for audit.
	terminal := req.Status == app.InstallStackVersionRunStatusSucceeded
	wantsOutputs := run.Kind == app.InstallStackVersionRunKindProvision ||
		run.Kind == app.InstallStackVersionRunKindReprovision
	if terminal && wantsOutputs {
		if err := s.dispatchUpdateInstallStackOutputs(ctx, &stackVersion, installID); err != nil {
			// Don't fail the run on dispatch error — the phone-home payload is
			// already saved, and a manual signal can replay it. Log loudly.
			s.l.Error("dispatch update install stack outputs",
				zap.Error(err),
				zap.String("install_id", installID),
				zap.String("run_id", runID))
		}
	}

	ctx.JSON(http.StatusOK, run)
}

// dispatchUpdateInstallStackOutputs fires the same workflow signal the legacy
// phone-home endpoint fires (see post_install_phone_home.go:146-174). The
// signal body is identical — only the trigger point differs.
func (s *service) dispatchUpdateInstallStackOutputs(ctx *gin.Context, stackVersion *app.InstallStackVersion, installID string) error {
	// Public endpoint: middleware doesn't set org or account on the context.
	// The features client needs the org; queue_signals INSERTs need an
	// account ID for the BeforeCreate hook's NOT NULL CreatedByID constraint.
	reqCtx := cctx.SetOrgIDContext(ctx.Request.Context(), stackVersion.OrgID)
	reqCtx = cctx.SetAccountIDContext(reqCtx, stackVersion.CreatedByID)

	useQueues, err := s.featuresClient.AllFeaturesEnabled(reqCtx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		return fmt.Errorf("checking features: %w", err)
	}
	if useQueues {
		queueID, err := s.getInstallSignalsQueueID(reqCtx, installID)
		if err != nil {
			return err
		}
		if err := s.enqueueInstallSignal(reqCtx, queueID, &updateinstallstackoutputs.Signal{
			InstallStackID: stackVersion.InstallStackID,
		}, "", ""); err != nil {
			return fmt.Errorf("enqueue signal: %w", err)
		}
		return nil
	}
	s.evClient.Send(reqCtx, installID, &signals.Signal{
		Type:           signals.OperationUpdateInstallStackOutputs,
		InstallStackID: stackVersion.InstallStackID,
	})
	return nil
}
