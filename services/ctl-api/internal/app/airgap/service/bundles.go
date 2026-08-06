package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type bundleResponse struct {
	ID                string                     `json:"id"`
	CreatedAt         time.Time                  `json:"created_at"`
	AppID             string                     `json:"app_id"`
	AppConfigID       string                     `json:"app_config_id"`
	TargetPlatform    string                     `json:"target_platform"`
	SchemaVersion     int                        `json:"schema_version"`
	ManifestDigest    string                     `json:"manifest_digest"`
	OCIRootDigest     string                     `json:"oci_root_digest"`
	TransportChecksum string                     `json:"transport_checksum"`
	Size              int64                      `json:"size"`
	Status            string                     `json:"status"`
	StatusDescription string                     `json:"status_description"`
	Artifacts         []app.AirgapBundleArtifact `json:"artifacts,omitempty"`
}

func responseFromBundle(bundle app.AirgapBundle) bundleResponse {
	return bundleResponse{
		ID: bundle.ID, CreatedAt: bundle.CreatedAt, AppID: bundle.AppID, AppConfigID: bundle.AppConfigID,
		TargetPlatform: bundle.TargetPlatform, SchemaVersion: bundle.SchemaVersion,
		ManifestDigest: bundle.ManifestDigest, OCIRootDigest: bundle.OCIRootDigest,
		TransportChecksum: bundle.TransportChecksum, Size: bundle.Size, Artifacts: bundle.Artifacts,
		Status: bundle.Status, StatusDescription: bundle.StatusDescription,
	}
}

// @ID ListAirgapBundles
// @Summary list published air-gap bundles for an app
// @Tags airgap-bundles
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param offset query int false "offset of results to return" Default(0)
// @Param limit query int false "limit of results to return" Default(10)
// @Param page query int false "page number of results to return" Default(0)
// @Success 200 {array} bundleResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/airgap-bundles [get]
func (s *service) ListBundles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	bundles, err := s.listBundles(ctx, org.ID, ctx.Param("app_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list air-gap bundles: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, bundles)
}

func (s *service) listBundles(ctx *gin.Context, orgID, appID string) ([]bundleResponse, error) {
	var bundles []app.AirgapBundle
	result := s.db.WithContext(ctx).Scopes(scopes.WithOffsetPagination).
		Where(app.AirgapBundle{OrgID: orgID, AppID: appID}).
		Order("created_at DESC").Find(&bundles)
	if result.Error != nil {
		return nil, result.Error
	}
	bundles, err := db.HandlePaginatedResponse(ctx, bundles)
	if err != nil {
		return nil, err
	}
	responses := make([]bundleResponse, len(bundles))
	for i, bundle := range bundles {
		responses[i] = responseFromBundle(bundle)
	}
	return responses, nil
}

// @ID GetAirgapBundle
// @Summary get a published air-gap bundle
// @Tags airgap-bundles
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param bundle_id path string true "bundle ID"
// @Success 200 {object} bundleResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/airgap-bundles/{bundle_id} [get]
func (s *service) GetBundle(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	bundle, err := s.getBundle(ctx, org.ID, ctx.Param("app_id"), ctx.Param("bundle_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get air-gap bundle: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, responseFromBundle(*bundle))
}

func (s *service) getBundle(ctx context.Context, orgID, appID, bundleID string) (*app.AirgapBundle, error) {
	var bundle app.AirgapBundle
	result := s.db.WithContext(ctx).Preload("Artifacts", func(db *gorm.DB) *gorm.DB {
		return db.Where(app.AirgapBundleArtifact{OrgID: orgID})
	}).
		Where(app.AirgapBundle{ID: bundleID, OrgID: orgID, AppID: appID}).First(&bundle)
	return &bundle, result.Error
}
