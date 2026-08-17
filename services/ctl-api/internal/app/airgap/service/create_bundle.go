package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	ocibundle "github.com/nuonco/nuon/pkg/runner/oci/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap"
	publishsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/signals/publish"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type createBundleRequest struct {
	AppConfigID    string                         `json:"app_config_id" binding:"required"`
	TargetPlatform string                         `json:"target_platform"`
	Runbooks       []runnerairgap.RunbookTemplate `json:"runbooks"`
}

// @ID CreateAirgapBundle
// @Summary create and publish an immutable air-gap bundle
// @Tags airgap-bundles
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param request body createBundleRequest true "bundle request"
// @Success 200 {object} bundleResponse
// @Success 202 {object} bundleResponse
// @Failure 400 {object} airgap.QualificationReport
// @Failure 412 {object} map[string]string
// @Router /v1/apps/{app_id}/airgap-bundles [post]
func (s *service) CreateBundle(ctx *gin.Context) {
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
			ctx.JSON(http.StatusBadRequest, report.report)
			return
		}
		if precondition, ok := err.(preconditionError); ok {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": precondition.msg})
			return
		}
		ctx.Error(fmt.Errorf("unable to create air-gap bundle: %w", err))
		return
	}
	ctx.JSON(status, responseFromBundle(*bundle))
}

type qualificationError struct{ report airgap.QualificationReport }

func (qualificationError) Error() string { return "app config does not qualify for air-gap export" }

type preconditionError struct{ msg string }

func (e preconditionError) Error() string { return e.msg }

func (s *service) createBundle(ctx context.Context, orgID, appID string, req createBundleRequest) (*app.AirgapBundle, int, error) {
	cfg, err := s.appsHelpers.GetAirgapAppConfig(ctx, orgID, appID, req.AppConfigID)
	if err != nil {
		return nil, 0, err
	}
	if report := airgap.Qualify(cfg, req.TargetPlatform); !report.Qualified {
		return nil, 0, qualificationError{report: report}
	}
	var bundle app.AirgapBundle
	err = s.db.WithContext(ctx).Preload("Replicas", func(db *gorm.DB) *gorm.DB { return db.Where(app.AirgapBundleTransportReplica{OrgID: orgID}) }).Where(app.AirgapBundle{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, TargetPlatform: req.TargetPlatform}).Order("created_at DESC").First(&bundle).Error
	if err == nil {
		for _, replica := range bundle.Replicas {
			if replica.VerifiedAt != nil && bundle.OCIIndexDigest != "" {
				if bundle.Status != app.AirgapBundleStatusActive {
					bundle.Status = app.AirgapBundleStatusActive
					bundle.StatusDescription = "bundle published and verified"
					if err := s.db.WithContext(ctx).Model(&bundle).Updates(app.AirgapBundle{Status: bundle.Status, StatusDescription: bundle.StatusDescription}).Error; err != nil {
						return nil, 0, err
					}
				}
				return &bundle, http.StatusOK, nil
			}
		}
		if bundle.SandboxBuildID == "" || len(bundle.ComponentBuildIDs) != len(cfg.ComponentConfigConnections) {
			selection, err := s.resolveActiveBuilds(ctx, orgID, appID, cfg)
			if err != nil {
				return nil, 0, err
			}
			bundle.SandboxBuildID = selection.sandboxBuildID
			bundle.ComponentBuildIDs = selection.componentBuildIDs
		}
		if bundle.Status != app.AirgapBundleStatusQueued && bundle.Status != app.AirgapBundleStatusPublishing {
			bundle.Status = app.AirgapBundleStatusQueued
			bundle.StatusDescription = "waiting to publish"
		}
		if err := s.db.WithContext(ctx).Model(&bundle).Updates(app.AirgapBundle{Status: bundle.Status, StatusDescription: bundle.StatusDescription, SandboxBuildID: bundle.SandboxBuildID, ComponentBuildIDs: bundle.ComponentBuildIDs}).Error; err != nil {
			return nil, 0, err
		}
	} else if err == gorm.ErrRecordNotFound {
		selection, err := s.resolveActiveBuilds(ctx, orgID, appID, cfg)
		if err != nil {
			return nil, 0, err
		}
		bundle = app.AirgapBundle{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, SandboxBuildID: selection.sandboxBuildID, ComponentBuildIDs: selection.componentBuildIDs, TargetPlatform: req.TargetPlatform, SchemaVersion: ocibundle.CurrentSchemaVersion, Status: app.AirgapBundleStatusQueued, StatusDescription: "waiting to publish"}
		if err := s.db.WithContext(ctx).Create(&bundle).Error; err != nil {
			return nil, 0, err
		}
	} else {
		return nil, 0, err
	}
	q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		return nil, 0, fmt.Errorf("get app queue: %w", err)
	}
	idempotencyKey := bundle.ID
	var latest app.QueueSignal
	err = s.db.WithContext(ctx).Where(app.QueueSignal{OwnerID: bundle.ID, OwnerType: "airgap_bundles", Type: publishsignal.SignalType}).Order("created_at DESC").First(&latest).Error
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
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{QueueID: q.ID, OwnerID: bundle.ID, OwnerType: "airgap_bundles", IdempotencyKey: idempotencyKey, Signal: &publishsignal.Signal{BundleID: bundle.ID, AppID: appID, Runbooks: req.Runbooks}})
	if err != nil {
		return nil, 0, fmt.Errorf("enqueue bundle publish signal: %w", err)
	}
	return &bundle, http.StatusAccepted, nil
}

type bundleBuildSelection struct {
	sandboxBuildID    string
	componentBuildIDs map[string]string
}

func (s *service) resolveActiveBuilds(ctx context.Context, orgID, appID string, cfg *app.AppConfig) (bundleBuildSelection, error) {
	selection := bundleBuildSelection{componentBuildIDs: make(map[string]string, len(cfg.ComponentConfigConnections))}
	for _, connection := range cfg.ComponentConfigConnections {
		var build app.ComponentBuild
		if err := s.db.WithContext(ctx).Where(app.ComponentBuild{OrgID: orgID, ComponentConfigConnectionID: connection.ID, Status: app.ComponentBuildStatusActive}).Order("created_at DESC").First(&build).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return selection, preconditionError{msg: fmt.Sprintf("no active build for component %s in app config %s", connection.ComponentName, cfg.ID)}
			}
			return selection, err
		}
		selection.componentBuildIDs[connection.ID] = build.ID
	}
	var sandboxBuild app.AppSandboxBuild
	if err := s.db.WithContext(ctx).Where(app.AppSandboxBuild{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive}).Order("created_at DESC").First(&sandboxBuild).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return selection, preconditionError{msg: fmt.Sprintf("no active sandbox build for app config %s: run a sandbox build for this config before creating a bundle", cfg.ID)}
		}
		return selection, err
	}
	selection.sandboxBuildID = sandboxBuild.ID
	return selection, nil
}
