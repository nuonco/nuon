package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type bundleResponse struct {
	ID                string     `json:"id"`
	CreatedAt         time.Time  `json:"created_at"`
	AppID             string     `json:"app_id"`
	AppConfigID       string     `json:"app_config_id"`
	TargetPlatform    string     `json:"target_platform"`
	SchemaVersion     int        `json:"schema_version"`
	ManifestDigest    string     `json:"manifest_digest"`
	OCIRootDigest     string     `json:"oci_root_digest"`
	OCIIndexDigest    string     `json:"oci_index_digest"`
	TransportChecksum string     `json:"transport_checksum"`
	Size              int64      `json:"size"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	Status            string     `json:"status"`
	StatusDescription string     `json:"status_description"`
}

func responseFromBundle(bundle app.AppBundle) bundleResponse {
	return bundleResponse{
		ID: bundle.ID, CreatedAt: bundle.CreatedAt, AppID: bundle.AppID, AppConfigID: bundle.AppConfigID,
		TargetPlatform: bundle.TargetPlatform, SchemaVersion: bundle.SchemaVersion,
		ManifestDigest: bundle.ManifestDigest, OCIRootDigest: bundle.OCIRootDigest, OCIIndexDigest: bundle.OCIIndexDigest,
		TransportChecksum: bundle.TransportChecksum, Size: bundle.Size, VerifiedAt: bundle.VerifiedAt,
		Status: string(bundle.Status), StatusDescription: bundle.StatusDescription,
	}
}

// @ID				GetAppBundle
// @Summary			get an app bundle
// @Description		Returns the bundle's publish status and any output metadata (manifest digests, archive checksum, size, verification timestamp) once publishing has completed.
// @Tags			app-bundles
// @Produce			json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id	path	string	true	"app ID"
// @Param			bundle_id	path	string	true	"bundle ID"
// @Success			200	{object}	bundleResponse
// @Failure			404	{object}	stderr.ErrResponse
// @Failure			500	{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/bundles/{bundle_id} [get]
func (s *service) GetBundle(ctx *gin.Context) {
	if !s.requireBundleExport(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	bundle, err := s.getBundle(ctx, org.ID, ctx.Param("app_id"), ctx.Param("bundle_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app bundle: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, responseFromBundle(*bundle))
}

func (s *service) getBundle(ctx context.Context, orgID, appID, bundleID string) (*app.AppBundle, error) {
	var bundle app.AppBundle
	result := s.db.WithContext(ctx).
		Where(app.AppBundle{ID: bundleID, OrgID: orgID, AppID: appID}).First(&bundle)
	return &bundle, result.Error
}
