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

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	ocibundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedapp "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	publishsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/signals/publish"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type createBundleRequest struct {
	AppConfigID    string                            `json:"app_config_id" binding:"required"`
	TargetPlatform string                            `json:"target_platform"`
	Runbooks       []customermanaged.RunbookTemplate `json:"runbooks"`
}

// @ID CreateCustomerManagedBundle
// @Summary create and publish an immutable portable bundle
// @Tags customer-managed-bundles
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param request body createBundleRequest true "bundle request"
// @Success 200 {object} bundleResponse
// @Success 202 {object} bundleResponse
// @Failure 400 {object} map[string]string
// @Failure 422 {object} customermanagedapp.QualificationReport
// @Failure 412 {object} map[string]string
// @Failure 500 {object} stderr.ErrResponse
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
			ctx.JSON(http.StatusUnprocessableEntity, report.report)
			return
		}
		if precondition, ok := err.(preconditionError); ok {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": precondition.msg})
			return
		}
		ctx.Error(fmt.Errorf("unable to create portable bundle: %w", err))
		return
	}
	ctx.JSON(status, responseFromBundle(*bundle))
}

type qualificationError struct {
	report customermanagedapp.QualificationReport
}

func (qualificationError) Error() string {
	return "app config does not qualify for portable bundle export"
}

type preconditionError struct{ msg string }

func (e preconditionError) Error() string { return e.msg }

func (s *service) createBundle(ctx context.Context, orgID, appID string, req createBundleRequest) (*app.CustomerManagedBundle, int, error) {
	cfg, err := s.appsHelpers.GetCustomerManagedAppConfig(ctx, orgID, appID, req.AppConfigID)
	if err != nil {
		return nil, 0, err
	}
	if report := customermanagedapp.Qualify(cfg, req.TargetPlatform); !report.Qualified {
		return nil, 0, qualificationError{report: report}
	}
	runbooks, runbooksDigest, err := canonicalBundleRunbooks(req.Runbooks)
	if err != nil {
		return nil, 0, fmt.Errorf("canonicalize bundle runbooks: %w", err)
	}
	var bundle app.CustomerManagedBundle
	err = s.db.WithContext(ctx).Preload("Replicas", func(db *gorm.DB) *gorm.DB { return db.Where(app.CustomerManagedBundleTransportReplica{OrgID: orgID}) }).Where(app.CustomerManagedBundle{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, TargetPlatform: req.TargetPlatform, RunbooksDigest: runbooksDigest}).Order("created_at DESC").First(&bundle).Error
	if err == nil {
		for _, replica := range bundle.Replicas {
			if replica.VerifiedAt != nil && bundle.OCIIndexDigest != "" {
				if bundle.Status != app.CustomerManagedBundleStatusActive {
					bundle.Status = app.CustomerManagedBundleStatusActive
					bundle.StatusDescription = "bundle published and verified"
					if err := s.db.WithContext(ctx).Model(&bundle).Updates(app.CustomerManagedBundle{Status: bundle.Status, StatusDescription: bundle.StatusDescription}).Error; err != nil {
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
		if bundle.Status != app.CustomerManagedBundleStatusQueued && bundle.Status != app.CustomerManagedBundleStatusPublishing {
			bundle.Status = app.CustomerManagedBundleStatusQueued
			bundle.StatusDescription = "waiting to publish"
		}
		if err := s.db.WithContext(ctx).Model(&bundle).Updates(app.CustomerManagedBundle{Status: bundle.Status, StatusDescription: bundle.StatusDescription, SandboxBuildID: bundle.SandboxBuildID, ComponentBuildIDs: bundle.ComponentBuildIDs}).Error; err != nil {
			return nil, 0, err
		}
	} else if err == gorm.ErrRecordNotFound {
		selection, err := s.resolveActiveBuilds(ctx, orgID, appID, cfg)
		if err != nil {
			return nil, 0, err
		}
		bundle = app.CustomerManagedBundle{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, SandboxBuildID: selection.sandboxBuildID, ComponentBuildIDs: selection.componentBuildIDs, Runbooks: runbooks, RunbooksDigest: runbooksDigest, TargetPlatform: req.TargetPlatform, SchemaVersion: ocibundle.CurrentSchemaVersion, Status: app.CustomerManagedBundleStatusQueued, StatusDescription: "waiting to publish"}
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
	err = s.db.WithContext(ctx).Where(app.QueueSignal{OwnerID: bundle.ID, OwnerType: "customer_managed_bundles", Type: publishsignal.SignalType}).Order("created_at DESC").First(&latest).Error
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
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{QueueID: q.ID, OwnerID: bundle.ID, OwnerType: "customer_managed_bundles", IdempotencyKey: idempotencyKey, Signal: &publishsignal.Signal{PackageID: bundle.ID, AppID: appID}})
	if err != nil {
		return nil, 0, fmt.Errorf("enqueue bundle publish signal: %w", err)
	}
	return &bundle, http.StatusAccepted, nil
}

func canonicalBundleRunbooks(runbooks []customermanaged.RunbookTemplate) ([]customermanaged.RunbookTemplate, string, error) {
	if len(runbooks) == 0 {
		return nil, "", nil
	}
	canonical := append([]customermanaged.RunbookTemplate(nil), runbooks...)
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
		if err := s.db.WithContext(ctx).Where(app.ComponentBuild{OrgID: orgID, ComponentConfigConnectionID: connection.ID, Status: app.ComponentBuildStatusActive}).Order("created_at DESC").First(&build).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return selection, preconditionError{msg: fmt.Sprintf("no active build for component %s in app config %s", connection.ComponentName, cfg.ID)}
			}
			return selection, err
		}
		selection.componentBuildIDs[connection.ID] = build.ID
	}
	sandboxBuild, found, err := s.resolveReleasedSandboxBuild(ctx, orgID, appID, cfg.SandboxConfig)
	if err != nil {
		return selection, err
	}
	if !found {
		err = s.db.WithContext(ctx).Where(app.AppSandboxBuild{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive}).Order("created_at DESC").First(&sandboxBuild).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return selection, preconditionError{msg: fmt.Sprintf("no active sandbox build for app config %s: run a sandbox build for this config before creating a bundle", cfg.ID)}
		}
		return selection, err
	}
	selection.sandboxBuildID = sandboxBuild.ID
	return selection, nil
}

func (s *service) resolveReleasedSandboxBuild(ctx context.Context, orgID, appID string, sandboxConfig app.AppSandboxConfig) (app.AppSandboxBuild, bool, error) {
	definition, err := customermanagedapp.CanonicalObject(sandboxConfig)
	if err != nil {
		return app.AppSandboxBuild{}, false, fmt.Errorf("canonicalize sandbox: %w", err)
	}
	var members []app.AppReleaseMember
	if err := s.db.WithContext(ctx).
		Preload("Release").
		Where(app.AppReleaseMember{OrgID: orgID, Kind: "sandbox", ConfigDigest: customermanagedapp.ObjectDigest(definition)}).
		Find(&members).Error; err != nil {
		return app.AppSandboxBuild{}, false, err
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].Release.CreatedAt.After(members[j].Release.CreatedAt)
	})
	for _, member := range members {
		if member.Release.AppID != appID || member.Release.Status != app.AppReleaseStatusReady {
			continue
		}
		var build app.AppSandboxBuild
		if err := s.db.WithContext(ctx).Where(app.AppSandboxBuild{
			ID: member.BuildID, OrgID: orgID, AppID: appID, Status: app.AppSandboxBuildStatusActive,
		}).First(&build).Error; err == nil {
			return build, true, nil
		} else if err != gorm.ErrRecordNotFound {
			return app.AppSandboxBuild{}, false, err
		}
	}
	return app.AppSandboxBuild{}, false, nil
}
