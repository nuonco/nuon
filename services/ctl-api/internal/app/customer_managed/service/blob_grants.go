package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const maxBlobGrantsPerRequest = 100

type blobGrantsRequest struct {
	// Digests are content-addressed blob digests (sha256 hex, with or
	// without the sha256: prefix) to grant download access for. When
	// empty, only bundle metadata is returned so clients can discover the
	// OCI index digest and diff against their local store first.
	Digests []string `json:"digests"`
}

type blobGrantItem struct {
	Digest    string    `json:"digest"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	ExpiresAt time.Time `json:"expires_at"`
}

type blobGrantsResponse struct {
	OCIIndexDigest    string          `json:"oci_index_digest"`
	ManifestDigest    string          `json:"manifest_digest"`
	TransportChecksum string          `json:"transport_checksum"`
	Grants            []blobGrantItem `json:"grants"`
}

// @ID CreateCustomerManagedBundleBlobGrants
// @Summary create download grants for individual content-addressed bundle blobs
// @Description Grants presigned access to individual bundle blobs so clients can download only blobs missing from their local store. Call with no digests to discover the bundle's OCI index digest, then request grants for missing blobs in batches.
// @Tags customer-managed-bundles
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param bundle_id path string true "bundle ID"
// @Param req body blobGrantsRequest true "blob grant request"
// @Success 200 {object} blobGrantsResponse
// @Failure 400 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/customer-managed-bundles/{bundle_id}/blob-grants [post]
func (s *service) CreateBlobGrants(ctx *gin.Context) {
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
	resp, err := s.createBlobGrants(ctx, org.ID, ctx.Param("app_id"), ctx.Param("bundle_id"), digests)
	if err != nil {
		if conflict, ok := err.(conflictError); ok {
			ctx.JSON(http.StatusConflict, gin.H{"error": conflict.Error()})
			return
		}
		ctx.Error(fmt.Errorf("unable to create portable bundle blob grants: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (s *service) createBlobGrants(ctx context.Context, orgID, appID, bundleID string, digests []string) (*blobGrantsResponse, error) {
	bundle, err := s.getBundle(ctx, orgID, appID, bundleID)
	if err != nil {
		return nil, err
	}
	if bundle.Status != app.CustomerManagedBundleStatusActive {
		return nil, conflictError{message: "portable bundle is not active"}
	}
	if bundle.OCIIndexDigest == "" {
		return nil, conflictError{message: "portable bundle was published before differential downloads were supported; republish the bundle to enable blob grants"}
	}
	indexDigest, err := canonicalDigest(bundle.OCIIndexDigest)
	if err != nil {
		return nil, fmt.Errorf("invalid stored OCI index digest: %w", err)
	}
	manifestDigest, err := canonicalDigest(bundle.ManifestDigest)
	if err != nil {
		return nil, fmt.Errorf("invalid stored manifest digest: %w", err)
	}
	resp := &blobGrantsResponse{
		OCIIndexDigest:    indexDigest,
		ManifestDigest:    manifestDigest,
		TransportChecksum: "sha256:" + strings.TrimPrefix(bundle.TransportChecksum, "sha256:"),
		Grants:            make([]blobGrantItem, 0, len(digests)),
	}
	// Grants are scoped to the org's content-addressed store rather than
	// validated against this bundle's blob closure: computing the closure
	// would require walking the OCI graph from object storage on every
	// request, and org members can already download any org bundle in full
	// via archive download grants, so blob-level access grants nothing new.
	for _, canonical := range digests {
		grant, err := s.store.GrantBlob(ctx, orgID, strings.TrimPrefix(canonical, "sha256:"))
		if err != nil {
			return nil, fmt.Errorf("grant blob %s: %w", canonical, err)
		}
		resp.Grants = append(resp.Grants, blobGrantItem{Digest: canonical, URL: grant.URL, Size: grant.Size, ExpiresAt: grant.ExpiresAt})
	}
	return resp, nil
}
