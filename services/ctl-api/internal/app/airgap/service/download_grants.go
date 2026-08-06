package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type downloadGrantResponse struct {
	URL               string    `json:"url"`
	ExpiresAt         time.Time `json:"expires_at"`
	Filename          string    `json:"filename"`
	Size              int64     `json:"size"`
	TransportChecksum string    `json:"transport_checksum"`
	ManifestDigest    string    `json:"manifest_digest"`
	SupportsRange     bool      `json:"supports_range"`
}

// @ID CreateAirgapBundleDownloadGrant
// @Summary create a download grant for a published air-gap bundle
// @Tags airgap-bundles
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param bundle_id path string true "bundle ID"
// @Success 200 {object} downloadGrantResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/airgap-bundles/{bundle_id}/download-grants [post]
func (s *service) CreateDownloadGrant(ctx *gin.Context) {
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
	grant, err := s.createDownloadGrant(ctx, org.ID, ctx.Param("app_id"), ctx.Param("bundle_id"))
	if err != nil {
		if conflict, ok := err.(conflictError); ok {
			ctx.JSON(http.StatusConflict, gin.H{"error": conflict.Error()})
			return
		}
		ctx.Error(fmt.Errorf("unable to create air-gap bundle download grant: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, grant)
}

func (s *service) createDownloadGrant(ctx context.Context, orgID, appID, bundleID string) (*downloadGrantResponse, error) {
	bundle, err := s.getBundle(ctx, orgID, appID, bundleID)
	if err != nil {
		return nil, err
	}
	if bundle.Status != app.AirgapBundleStatusActive {
		return nil, conflictError{message: "air-gap bundle is not active or does not have a verified replica"}
	}
	var replicas []app.AirgapBundleTransportReplica
	result := s.db.WithContext(ctx).
		Where(app.AirgapBundleTransportReplica{OrgID: orgID, BundleID: bundle.ID}).
		Order("verified_at DESC").Order("created_at DESC").Order("id ASC").Find(&replicas)
	if result.Error != nil {
		return nil, result.Error
	}
	var replica *app.AirgapBundleTransportReplica
	for i := range replicas {
		if replicas[i].VerifiedAt != nil {
			replica = &replicas[i]
			break
		}
	}
	if replica == nil {
		return nil, conflictError{message: "air-gap bundle is not active or does not have a verified replica"}
	}
	manifestDigest, err := canonicalDigest(bundle.ManifestDigest)
	if err != nil {
		return nil, fmt.Errorf("invalid stored manifest digest: %w", err)
	}
	transportChecksum, err := canonicalDigest(replica.TransportChecksum)
	if err != nil {
		return nil, fmt.Errorf("invalid stored transport checksum: %w", err)
	}
	filename := "app-bundle-" + strings.TrimPrefix(manifestDigest, "sha256:")[:12] + ".oci.tar.zst"
	storeGrant, err := s.store.Grant(ctx, transport.Replica{
		Provider: replica.Provider, Region: replica.Region, StorageRef: replica.StorageRef,
		StorageVersion: replica.StorageVersion, TransportChecksum: replica.TransportChecksum,
		Size: replica.Size, VerifiedAt: *replica.VerifiedAt,
	}, filename, time.Time{})
	if err != nil {
		return nil, err
	}
	return &downloadGrantResponse{
		URL: storeGrant.URL, ExpiresAt: storeGrant.ExpiresAt, Filename: filename, Size: replica.Size,
		TransportChecksum: transportChecksum, ManifestDigest: manifestDigest, SupportsRange: storeGrant.SupportsRange,
	}, nil
}

type conflictError struct{ message string }

func (e conflictError) Error() string { return e.message }

func canonicalDigest(value string) (string, error) {
	value = strings.TrimPrefix(strings.ToLower(value), "sha256:")
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return "", errors.New("SHA-256 must be 64 hexadecimal characters")
	}
	return "sha256:" + value, nil
}
