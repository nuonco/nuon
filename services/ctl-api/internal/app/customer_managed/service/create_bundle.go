package service

import (
	"context"
	"fmt"
	"net/http"

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
// @Router /v1/apps/{app_id}/customer-managed-bundles [post]
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
