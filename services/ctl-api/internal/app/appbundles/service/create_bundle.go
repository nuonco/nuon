package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	appbundle "github.com/nuonco/nuon/pkg/appbundle"
	ocibundle "github.com/nuonco/nuon/pkg/appbundle/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appbundles "github.com/nuonco/nuon/services/ctl-api/internal/app/appbundles"
	publishsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/appbundles/signals/publish"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/appbundles/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

type createBundleRequest struct {
	AppConfigID    string                      `json:"app_config_id" binding:"required"`
	TargetPlatform string                      `json:"target_platform"`
	Runbooks       []appbundle.RunbookTemplate `json:"runbooks"`
	Runtime        *app.AppBundleRuntime       `json:"runtime"`
}

// @ID				CreateAppBundle
// @Summary			create and publish an immutable app bundle from an app config
// @Description		Resolves the app config's successful sandbox and component builds once, pins them on the bundle, and enqueues an asynchronous publish that assembles an OCI-layout .tar.zst archive. Retries reuse the pinned builds and never select newer builds.
// @Tags			app-bundles
// @Accept			json
// @Produce			json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id	path	string	true	"app ID"
// @Param			request	body	createBundleRequest	true	"bundle request"
// @Success			200	{object}	bundleResponse
// @Success			202	{object}	bundleResponse
// @Failure			400	{object}	stderr.ErrResponse
// @Failure			404	{object}	stderr.ErrResponse
// @Failure			412	{object}	stderr.ErrResponse
// @Failure			422	{object}	appbundles.QualificationReport
// @Failure			500	{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/bundles [post]
func (s *service) CreateBundle(ctx *gin.Context) {
	if !s.requireBundleExport(ctx) {
		return
	}
	if !s.store.Configured() {
		ctx.Error(transport.ErrNotConfigured)
		return
	}
	var req createBundleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if req.TargetPlatform == "" {
		req.TargetPlatform = "linux/amd64"
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	bundle, status, err := s.createBundle(ctx, org.ID, ctx.Param("app_id"), req)
	if err != nil {
		if report, ok := err.(qualificationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, report.report)
			return
		}
		if precondition, ok := err.(preconditionError); ok {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": precondition.msg})
			return
		}
		ctx.Error(fmt.Errorf("unable to create app bundle: %w", err))
		return
	}
	ctx.JSON(status, responseFromBundle(*bundle))
}

type qualificationError struct {
	report appbundles.QualificationReport
}

func (qualificationError) Error() string {
	return "app config does not qualify for bundle export"
}

type preconditionError struct{ msg string }

func (e preconditionError) Error() string { return e.msg }

func (s *service) createBundle(ctx context.Context, orgID, appID string, req createBundleRequest) (*app.AppBundle, int, error) {
	cfg, err := s.appsHelpers.GetAppBundleAppConfig(ctx, orgID, appID, req.AppConfigID)
	if err != nil {
		return nil, 0, err
	}
	if report := appbundles.Qualify(cfg, req.TargetPlatform); !report.Qualified {
		return nil, 0, qualificationError{report: report}
	}
	runtime := app.AppBundleRuntime{}
	if req.Runtime != nil {
		if err := validateBundleRuntime(*req.Runtime); err != nil {
			return nil, 0, preconditionError{msg: err.Error()}
		}
		runtime = *req.Runtime
	}
	runtimeDigest := appbundles.ObjectDigest(runtime)
	runbooks, runbooksDigest, err := canonicalBundleRunbooks(req.Runbooks)
	if err != nil {
		return nil, 0, fmt.Errorf("canonicalize bundle runbooks: %w", err)
	}
	var bundle app.AppBundle
	err = s.db.WithContext(ctx).
		// Map-based Where, not a struct: GORM skips zero-valued struct fields,
		// so an empty RunbooksDigest/RuntimeDigest would match a bundle that
		// has runbooks or runtime pins.
		Where(map[string]any{"org_id": orgID, "app_id": appID, "app_config_id": cfg.ID, "target_platform": req.TargetPlatform, "runbooks_digest": runbooksDigest, "runtime_digest": runtimeDigest}).
		Order("created_at DESC").First(&bundle).Error
	if err == nil {
		if bundle.HasVerifiedArchive() {
			if bundle.Status != app.AppBundleStatusActive {
				bundle.Status = app.AppBundleStatusActive
				bundle.StatusDescription = "bundle published and verified"
				if err := s.db.WithContext(ctx).Model(&bundle).Updates(app.AppBundle{Status: bundle.Status, StatusDescription: bundle.StatusDescription}).Error; err != nil {
					return nil, 0, err
				}
			}
			return &bundle, http.StatusOK, nil
		}
		if bundle.SandboxBuildID == "" || len(bundle.ComponentBuildIDs) != len(cfg.ComponentConfigConnections) {
			selection, err := s.resolveActiveBuilds(ctx, orgID, appID, cfg)
			if err != nil {
				return nil, 0, err
			}
			bundle.SandboxBuildID = selection.sandboxBuildID
			bundle.ComponentBuildIDs = selection.componentBuildIDs
		}
		if bundle.Status != app.AppBundleStatusQueued && bundle.Status != app.AppBundleStatusPublishing {
			bundle.Status = app.AppBundleStatusQueued
			bundle.StatusDescription = "waiting to publish"
		}
		if err := s.db.WithContext(ctx).Model(&bundle).Updates(app.AppBundle{Status: bundle.Status, StatusDescription: bundle.StatusDescription, SandboxBuildID: bundle.SandboxBuildID, ComponentBuildIDs: bundle.ComponentBuildIDs}).Error; err != nil {
			return nil, 0, err
		}
	} else if err == gorm.ErrRecordNotFound {
		selection, err := s.resolveActiveBuilds(ctx, orgID, appID, cfg)
		if err != nil {
			return nil, 0, err
		}
		bundle = app.AppBundle{
			OrgID: orgID, AppID: appID, AppConfigID: cfg.ID,
			SandboxBuildID: selection.sandboxBuildID, ComponentBuildIDs: selection.componentBuildIDs,
			Runbooks: runbooks, RunbooksDigest: runbooksDigest,
			Runtime: runtime, RuntimeDigest: runtimeDigest,
			TargetPlatform: req.TargetPlatform, SchemaVersion: ocibundle.CurrentSchemaVersion,
			Status: app.AppBundleStatusQueued, StatusDescription: "waiting to publish",
		}
		if err := s.db.WithContext(ctx).Create(&bundle).Error; err != nil {
			return nil, 0, err
		}
	} else {
		return nil, 0, err
	}
	q, err := s.queueClient.GetQueueByOwnerAndName(ctx, appID, queuenames.OwnerApps, queuenames.AppSignalsQueueName)
	if err != nil {
		return nil, 0, fmt.Errorf("get app queue: %w", err)
	}
	idempotencyKey := bundle.ID
	var latest app.QueueSignal
	err = s.db.WithContext(ctx).Where(app.QueueSignal{OwnerID: bundle.ID, OwnerType: "app_bundles", Type: publishsignal.SignalType}).Order("created_at DESC").First(&latest).Error
	if err == nil {
		switch latest.Status.Status {
		case app.StatusSuccess, app.StatusError, app.StatusCancelled:
			idempotencyKey += ":" + latest.ID
		default:
			return &bundle, http.StatusAccepted, nil
		}
	} else if err != gorm.ErrRecordNotFound {
		return nil, 0, fmt.Errorf("load bundle publish signal: %w", err)
	}
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{QueueID: q.ID, OwnerID: bundle.ID, OwnerType: "app_bundles", IdempotencyKey: idempotencyKey, Signal: &publishsignal.Signal{BundleID: bundle.ID, AppID: appID}})
	if err != nil {
		return nil, 0, fmt.Errorf("enqueue bundle publish signal: %w", err)
	}
	return &bundle, http.StatusAccepted, nil
}

func validateBundleRuntime(runtime app.AppBundleRuntime) error {
	if runtime.RunnerImageURL == "" || runtime.RunnerImageTag == "" {
		return fmt.Errorf("runner_image_url and runner_image_tag are required")
	}
	platform, ok := runtime.Platforms["linux/amd64"]
	if !ok {
		return fmt.Errorf("platform linux/amd64 is required")
	}
	if platform.RunnerBinaryURL == "" {
		return fmt.Errorf("platform linux/amd64 requires runner_binary_url")
	}
	return nil
}

func canonicalBundleRunbooks(runbooks []appbundle.RunbookTemplate) ([]appbundle.RunbookTemplate, string, error) {
	if len(runbooks) == 0 {
		return nil, "", nil
	}
	canonical := append([]appbundle.RunbookTemplate(nil), runbooks...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].ID == canonical[j].ID {
			return canonical[i].Name < canonical[j].Name
		}
		return canonical[i].ID < canonical[j].ID
	})
	raw, err := json.Marshal(canonical)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(raw)
	return canonical, hex.EncodeToString(digest[:]), nil
}

type bundleBuildSelection struct {
	sandboxBuildID    string
	componentBuildIDs map[string]string
}

func (s *service) resolveActiveBuilds(ctx context.Context, orgID, appID string, cfg *app.AppConfig) (bundleBuildSelection, error) {
	selection := bundleBuildSelection{componentBuildIDs: make(map[string]string, len(cfg.ComponentConfigConnections))}
	for _, connection := range cfg.ComponentConfigConnections {
		var build app.ComponentBuild
		var err error
		if connection.LatestBuildID.Valid {
			build, err = s.activeBuildForConnection(ctx, orgID, connection.LatestBuildID.String, connection)
		} else {
			err = s.db.WithContext(ctx).Where(app.ComponentBuild{OrgID: orgID, ComponentConfigConnectionID: connection.ID, Status: app.ComponentBuildStatusActive}).Order("created_at DESC").First(&build).Error
		}
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return selection, preconditionError{msg: fmt.Sprintf("no active build for component %s in app config %s", connection.ComponentName, cfg.ID)}
			}
			return selection, err
		}
		selection.componentBuildIDs[connection.ID] = build.ID
	}
	var sandboxBuild app.AppSandboxBuild
	err := s.db.WithContext(ctx).Where(app.AppSandboxBuild{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive}).Order("created_at DESC").First(&sandboxBuild).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return selection, preconditionError{msg: fmt.Sprintf("no active sandbox build for app config %s: run a sandbox build for this config before creating a bundle", cfg.ID)}
		}
		return selection, err
	}
	selection.sandboxBuildID = sandboxBuild.ID
	return selection, nil
}

func (s *service) activeBuildForConnection(ctx context.Context, orgID, buildID string, connection app.ComponentConfigConnection) (app.ComponentBuild, error) {
	var build app.ComponentBuild
	if err := s.db.WithContext(ctx).
		Where(app.ComponentBuild{ID: buildID, OrgID: orgID, Status: app.ComponentBuildStatusActive}).
		First(&build).Error; err != nil {
		return build, err
	}
	var buildComponentID string
	if err := s.db.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where(app.ComponentConfigConnection{ID: build.ComponentConfigConnectionID, OrgID: orgID}).
		Pluck("component_id", &buildComponentID).Error; err != nil {
		return build, err
	}
	if buildComponentID != connection.ComponentID {
		return build, gorm.ErrRecordNotFound
	}
	return build, nil
}
