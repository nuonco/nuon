package service

import (
	"context"
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

	// Legacy route omits the kind segment — default to provision so older SDK
	// builds (and any existing CLI sessions mid-flight during rollout) keep
	// working.
	kind := app.InstallStackVersionRunKind(ctx.Param("kind"))
	if kind == "" {
		kind = app.InstallStackVersionRunKindProvision
	}
	if !kind.Valid() {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("unknown stack run kind: %q", kind)))
		return
	}

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
		Kind:                  kind,
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

	parentStatus := app.InstallStackVersionStatusProvisioning
	if kind == app.InstallStackVersionRunKindDeprovision {
		parentStatus = app.InstallStackVersionStatusDestroying
	}
	if res := s.db.WithContext(reqCtx).
		Model(&app.InstallStackVersion{ID: stackVersion.ID}).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(reqCtx, parentStatus),
		}); res.Error != nil {
		ctx.Error(fmt.Errorf("update stack version status: %w", res.Error))
		return
	}

	// Build the SDK config block. A missing config is fatal — the runner can't
	// bootstrap without runner_id + runner_api_url, and silently shipping an
	// empty config would leave EC2 instances tagged with garbage that fails
	// authentication minutes later when the dashboard expects them online.
	cfg, err := s.buildInstallerSDKConfig(reqCtx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("build installer sdk config: %w", err))
		return
	}
	run.SDKConfig = cfg

	ctx.JSON(http.StatusCreated, run)
}

// buildInstallerSDKConfig assembles the SDK config block from the install +
// runner records. Sources runner_api_url from the per-runner-group settings
// (same as the CFN renderer at pkg/stacks/cloudformation/resource_ec2_launch_template.go),
// falling back to the global ctl-api config only when the per-group value is
// missing. Operation roles, secrets, and dynamic role configs are
// intentionally left empty for now — the SDK handles empty/nil sub-configs
// as no-ops; full tfvars-equivalent rendering is tracked separately.
func (s *service) buildInstallerSDKConfig(ctx context.Context, installID string) (*app.InstallerSDKConfig, error) {
	var install app.Install
	if res := s.db.WithContext(ctx).
		Preload("RunnerGroup.Runners").
		Preload("RunnerGroup.Settings").
		Where("id = ?", installID).
		First(&install); res.Error != nil {
		return nil, fmt.Errorf("load install: %w", res.Error)
	}

	// Per-runner-group setting wins; global config is only a safety net for
	// older installs that pre-date the per-group field being populated.
	runnerAPIURL := install.RunnerGroup.Settings.RunnerAPIURL
	if runnerAPIURL == "" {
		runnerAPIURL = s.cfg.RunnerAPIURL
	}

	if install.RunnerID == "" {
		return nil, fmt.Errorf("install %s has no runner — cannot build SDK config", installID)
	}
	if runnerAPIURL == "" {
		return nil, fmt.Errorf("install %s: runner_api_url empty (set RunnerGroupSettings.RunnerAPIURL for the install's runner group)", installID)
	}

	return &app.InstallerSDKConfig{
		InstallID:    install.ID,
		RunnerID:     install.RunnerID,
		RunnerAPIURL: runnerAPIURL,
	}, nil
}
