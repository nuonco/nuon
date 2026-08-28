package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanaged "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	publishsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/signals/publish"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type createReleasePackageRequest struct {
	Format         string `json:"format"`
	TargetPlatform string `json:"target_platform"`
}

// @ID CreateReleasePackage
// @Summary create a portable package for an application release
// @Tags release-packages
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param release_id path string true "release ID"
// @Param request body createReleasePackageRequest true "package request"
// @Success 200 {object} app.ReleasePackage
// @Success 202 {object} app.ReleasePackage
// @Failure 400 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 422 {object} customermanaged.QualificationReport
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases/{release_id}/packages [post]
func (s *service) CreateReleasePackage(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	if !s.store.Configured() {
		ctx.Error(transport.ErrNotConfigured)
		return
	}
	var req createReleasePackageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if req.Format == "" {
		req.Format = app.ReleasePackageFormatPortableOCI
	}
	if req.TargetPlatform == "" {
		req.TargetPlatform = "linux/amd64"
	}
	if req.Format != app.ReleasePackageFormatPortableOCI {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported package format %q", req.Format)})
		return
	}
	if _, err := bundleTarget(req.TargetPlatform); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	pkg, status, err := s.createReleasePackage(ctx, org.ID, ctx.Param("app_id"), ctx.Param("release_id"), req)
	if err != nil {
		if report, ok := err.(qualificationError); ok {
			ctx.JSON(http.StatusUnprocessableEntity, report.report)
			return
		}
		ctx.Error(fmt.Errorf("unable to create release package: %w", err))
		return
	}
	ctx.JSON(status, pkg)
}

func (s *service) createReleasePackage(ctx context.Context, orgID, appID, releaseID string, req createReleasePackageRequest) (*app.ReleasePackage, int, error) {
	var release app.AppRelease
	if err := s.db.WithContext(ctx).Where(app.AppRelease{ID: releaseID, OrgID: orgID, AppID: appID}).First(&release).Error; err != nil {
		return nil, 0, err
	}
	cfg, err := s.appsHelpers.GetCustomerManagedAppConfig(ctx, orgID, appID, release.AppConfigID)
	if err != nil {
		return nil, 0, err
	}
	if report := customermanaged.Qualify(cfg, req.TargetPlatform); !report.Qualified {
		return nil, 0, qualificationError{report: report}
	}
	packageDigest := releasePackageDigest(release.SemanticDigest, release.RuntimeDigest, req)
	var pkg app.ReleasePackage
	err = s.db.WithContext(ctx).
		Preload("Replicas", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackageReplica{OrgID: orgID}) }).
		Where(app.ReleasePackage{OrgID: orgID, ReleaseID: release.ID, Format: req.Format, TargetPlatform: req.TargetPlatform}).First(&pkg).Error
	if err == nil {
		if pkg.Status == app.ReleasePackageStatusActive {
			return &pkg, http.StatusOK, nil
		}
		if pkg.Status == app.ReleasePackageStatusPublishing || pkg.Status == app.ReleasePackageStatusQueued {
			return &pkg, http.StatusAccepted, nil
		}
		pkg.Status = app.ReleasePackageStatusQueued
		pkg.StatusDescription = "waiting to publish"
		if err := s.db.WithContext(ctx).Model(&pkg).Updates(app.ReleasePackage{Status: pkg.Status, StatusDescription: pkg.StatusDescription}).Error; err != nil {
			return nil, 0, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		pkg = app.ReleasePackage{
			OrgID: orgID, ReleaseID: release.ID, Format: req.Format, TargetPlatform: req.TargetPlatform,
			PackageDigest: packageDigest, SchemaVersion: bundle.CurrentSchemaVersion,
			Status: app.ReleasePackageStatusQueued, StatusDescription: "waiting to publish",
		}
		if err := s.db.WithContext(ctx).Create(&pkg).Error; err != nil {
			return nil, 0, err
		}
	} else {
		return nil, 0, err
	}
	queue, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		return nil, 0, fmt.Errorf("get app queue: %w", err)
	}
	idempotencyKey := pkg.ID
	var latest app.QueueSignal
	err = s.db.WithContext(ctx).Where(app.QueueSignal{OwnerID: pkg.ID, OwnerType: "release_packages", Type: publishsignal.SignalType}).Order("created_at DESC").First(&latest).Error
	if err == nil {
		switch latest.Status.Status {
		case app.StatusSuccess, app.StatusError, app.StatusCancelled:
			idempotencyKey += ":" + latest.ID
		default:
			return &pkg, http.StatusAccepted, nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, fmt.Errorf("load package publish signal: %w", err)
	}
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID, OwnerID: pkg.ID, OwnerType: "release_packages", IdempotencyKey: idempotencyKey,
		Signal: &publishsignal.Signal{PackageID: pkg.ID, AppID: appID},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("enqueue package publish signal: %w", err)
	}
	return &pkg, http.StatusAccepted, nil
}

func releasePackageDigest(releaseDigest, runtimeDigest string, req createReleasePackageRequest) string {
	return customermanaged.ObjectDigest(map[string]any{
		"release_digest": releaseDigest, "format": req.Format, "target_platform": req.TargetPlatform,
		"runtime_digest": runtimeDigest,
	})
}

// @ID ListReleasePackages
// @Summary list packages for an application release
// @Tags release-packages
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param release_id path string true "release ID"
// @Success 200 {array} app.ReleasePackage
// @Failure 403 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases/{release_id}/packages [get]
func (s *service) ListReleasePackages(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var release app.AppRelease
	if err := s.db.WithContext(ctx).Where(app.AppRelease{ID: ctx.Param("release_id"), OrgID: org.ID, AppID: ctx.Param("app_id")}).First(&release).Error; err != nil {
		ctx.Error(err)
		return
	}
	var packages []app.ReleasePackage
	if err := s.db.WithContext(ctx).Where(app.ReleasePackage{OrgID: org.ID, ReleaseID: release.ID}).Order("created_at DESC").Find(&packages).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, packages)
}

// @ID GetReleasePackage
// @Summary get a release package
// @Tags release-packages
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param package_id path string true "package ID"
// @Success 200 {object} app.ReleasePackage
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/release-packages/{package_id} [get]
func (s *service) GetReleasePackage(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	pkg, err := s.getReleasePackage(ctx, org.ID, ctx.Param("package_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, pkg)
}

func (s *service) getReleasePackage(ctx context.Context, orgID, packageID string) (*app.ReleasePackage, error) {
	var pkg app.ReleasePackage
	err := s.db.WithContext(ctx).
		Preload("Release", func(db *gorm.DB) *gorm.DB { return db.Where(app.AppRelease{OrgID: orgID}) }).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackageMember{OrgID: orgID}) }).
		Preload("Replicas", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackageReplica{OrgID: orgID}) }).
		Where(app.ReleasePackage{ID: packageID, OrgID: orgID}).First(&pkg).Error
	return &pkg, err
}

type releasePackageDownloadGrantResponse struct {
	URL             string    `json:"url"`
	ExpiresAt       time.Time `json:"expires_at"`
	Filename        string    `json:"filename"`
	Size            int64     `json:"size"`
	ArchiveChecksum string    `json:"archive_checksum"`
	ManifestDigest  string    `json:"manifest_digest"`
	SupportsRange   bool      `json:"supports_range"`
}

// @ID CreateReleasePackageDownloadGrant
// @Summary create a download grant for a published release package
// @Tags release-packages
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param package_id path string true "package ID"
// @Success 200 {object} releasePackageDownloadGrantResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/release-packages/{package_id}/download-grants [post]
func (s *service) CreateReleasePackageDownloadGrant(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	if !s.store.Configured() {
		ctx.Error(transport.ErrNotConfigured)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	grant, err := s.createReleasePackageDownloadGrant(ctx, org.ID, ctx.Param("package_id"))
	if err != nil {
		if conflict, ok := err.(conflictError); ok {
			ctx.JSON(http.StatusConflict, gin.H{"error": conflict.Error()})
			return
		}
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, grant)
}

func (s *service) createReleasePackageDownloadGrant(ctx context.Context, orgID, packageID string) (*releasePackageDownloadGrantResponse, error) {
	pkg, err := s.getReleasePackage(ctx, orgID, packageID)
	if err != nil {
		return nil, err
	}
	if pkg.Status != app.ReleasePackageStatusActive {
		return nil, conflictError{message: "release package is not active"}
	}
	var replicas []app.ReleasePackageReplica
	if err := s.db.WithContext(ctx).Where(app.ReleasePackageReplica{OrgID: orgID, PackageID: pkg.ID}).Order("verified_at DESC").Order("created_at DESC").Find(&replicas).Error; err != nil {
		return nil, err
	}
	var replica *app.ReleasePackageReplica
	for i := range replicas {
		if replicas[i].VerifiedAt != nil {
			replica = &replicas[i]
			break
		}
	}
	if replica == nil {
		return nil, conflictError{message: "release package does not have a verified replica"}
	}
	manifestDigest, err := canonicalDigest(pkg.ManifestDigest)
	if err != nil {
		return nil, fmt.Errorf("invalid stored manifest digest: %w", err)
	}
	archiveChecksum, err := canonicalDigest(replica.ArchiveChecksum)
	if err != nil {
		return nil, fmt.Errorf("invalid stored archive checksum: %w", err)
	}
	filename := "app-bundle-" + strings.TrimPrefix(manifestDigest, "sha256:")[:12] + ".oci.tar.zst"
	storeGrant, err := s.store.Grant(ctx, transport.Replica{
		Provider: replica.Provider, Region: replica.Region, StorageRef: replica.StorageRef,
		StorageVersion: replica.StorageVersion, TransportChecksum: replica.ArchiveChecksum,
		Size: replica.Size, VerifiedAt: *replica.VerifiedAt,
	}, filename, time.Time{})
	if err != nil {
		return nil, err
	}
	return &releasePackageDownloadGrantResponse{
		URL: storeGrant.URL, ExpiresAt: storeGrant.ExpiresAt, Filename: filename, Size: replica.Size,
		ArchiveChecksum: archiveChecksum, ManifestDigest: manifestDigest, SupportsRange: storeGrant.SupportsRange,
	}, nil
}

// @ID CreateReleasePackageBlobGrants
// @Summary create download grants for content-addressed package blobs
// @Tags release-packages
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param package_id path string true "package ID"
// @Param request body blobGrantsRequest true "blob grant request"
// @Success 200 {object} blobGrantsResponse
// @Failure 400 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/release-packages/{package_id}/blob-grants [post]
func (s *service) CreateReleasePackageBlobGrants(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	if !s.store.Configured() {
		ctx.Error(transport.ErrNotConfigured)
		return
	}
	var req blobGrantsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(req.Digests) > maxBlobGrantsPerRequest {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("at most %d blob grants can be requested at once", maxBlobGrantsPerRequest)})
		return
	}
	digests := make([]string, 0, len(req.Digests))
	for _, requested := range req.Digests {
		canonical, err := canonicalDigest(requested)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid blob digest %q", requested)})
			return
		}
		digests = append(digests, canonical)
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	resp, err := s.createReleasePackageBlobGrants(ctx, org.ID, ctx.Param("package_id"), digests)
	if err != nil {
		if conflict, ok := err.(conflictError); ok {
			ctx.JSON(http.StatusConflict, gin.H{"error": conflict.Error()})
			return
		}
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (s *service) createReleasePackageBlobGrants(ctx context.Context, orgID, packageID string, digests []string) (*blobGrantsResponse, error) {
	pkg, err := s.getReleasePackage(ctx, orgID, packageID)
	if err != nil {
		return nil, err
	}
	if pkg.Status != app.ReleasePackageStatusActive || pkg.OCIIndexDigest == "" {
		return nil, conflictError{message: "release package is not active"}
	}
	indexDigest, err := canonicalDigest(pkg.OCIIndexDigest)
	if err != nil {
		return nil, fmt.Errorf("invalid stored OCI index digest: %w", err)
	}
	manifestDigest, err := canonicalDigest(pkg.ManifestDigest)
	if err != nil {
		return nil, fmt.Errorf("invalid stored manifest digest: %w", err)
	}
	response := &blobGrantsResponse{
		OCIIndexDigest: indexDigest, ManifestDigest: manifestDigest,
		TransportChecksum: "sha256:" + strings.TrimPrefix(pkg.ArchiveChecksum, "sha256:"),
		Grants:            make([]blobGrantItem, 0, len(digests)),
	}
	for _, canonical := range digests {
		grant, err := s.store.GrantBlob(ctx, orgID, strings.TrimPrefix(canonical, "sha256:"))
		if err != nil {
			return nil, fmt.Errorf("grant blob %s: %w", canonical, err)
		}
		response.Grants = append(response.Grants, blobGrantItem{Digest: canonical, URL: grant.URL, Size: grant.Size, ExpiresAt: grant.ExpiresAt})
	}
	return response, nil
}

func bundleTarget(targetPlatform string) (bundle.Target, error) {
	parts := strings.Split(targetPlatform, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return bundle.Target{}, fmt.Errorf("invalid target platform %q: expected os/architecture", targetPlatform)
	}
	return bundle.Target{OS: parts[0], Architecture: parts[1]}, nil
}
