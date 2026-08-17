package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestCreateBlobGrantsReturnsMetadataAndGrants(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, store := testService(t)
	now := time.Now()
	digest := strings.Repeat("a", 64)
	indexDigest := strings.Repeat("b", 64)
	blobDigest := strings.Repeat("c", 64)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)
	require.NoError(t, svc.db.Model(&app.AirgapBundle{ID: bundle.ID}).Update("oci_index_digest", "sha256:"+indexDigest).Error)
	store.blobGrants = map[string]transport.BlobGrant{
		blobDigest: {URL: "https://example.test/blob", Size: 7, ExpiresAt: now.Add(time.Minute)},
	}

	resp, err := svc.createBlobGrants(context.Background(), "org-a", "app-a", bundle.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "sha256:"+indexDigest, resp.OCIIndexDigest)
	require.Equal(t, "sha256:"+digest, resp.ManifestDigest)
	require.Equal(t, "sha256:"+digest, resp.TransportChecksum)
	require.Empty(t, resp.Grants)

	resp, err = svc.createBlobGrants(context.Background(), "org-a", "app-a", bundle.ID, []string{"sha256:" + blobDigest})
	require.NoError(t, err)
	require.Len(t, resp.Grants, 1)
	require.Equal(t, "sha256:"+blobDigest, resp.Grants[0].Digest)
	require.Equal(t, "https://example.test/blob", resp.Grants[0].URL)
	require.Equal(t, int64(7), resp.Grants[0].Size)
	require.Equal(t, []string{blobDigest}, store.granted)
}

func TestCreateBlobGrantsRequiresIndexDigest(t *testing.T) {
	svc, _ := testService(t)
	now := time.Now()
	digest := strings.Repeat("a", 64)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)

	_, err := svc.createBlobGrants(context.Background(), "org-a", "app-a", bundle.ID, nil)
	require.ErrorContains(t, err, "republish the bundle")
}

func TestCreateBlobGrantsHandlerValidatesRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, _ := testService(t)
	now := time.Now()
	digest := strings.Repeat("a", 64)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)
	require.NoError(t, svc.db.Model(&app.AirgapBundle{ID: bundle.ID}).Update("oci_index_digest", "sha256:"+digest).Error)

	post := func(body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("POST", "/", strings.NewReader(body))
		ctx.Params = gin.Params{{Key: "app_id", Value: "app-a"}, {Key: "bundle_id", Value: bundle.ID}}
		cctx.SetOrgGinContext(ctx, &app.Org{ID: "org-a"})
		svc.CreateBlobGrants(ctx)
		return recorder
	}

	require.Equal(t, 400, post(`{"digests":["not-a-digest"]}`).Code)
	require.Equal(t, 400, post(`{`).Code)

	tooMany := make([]string, maxBlobGrantsPerRequest+1)
	for i := range tooMany {
		tooMany[i] = "sha256:" + strings.Repeat("d", 64)
	}
	encoded, err := json.Marshal(map[string][]string{"digests": tooMany})
	require.NoError(t, err)
	require.Equal(t, 400, post(string(encoded)).Code)

	metadata := post(`{"digests":[]}`)
	require.Equal(t, 200, metadata.Code)
	require.Equal(t, "no-store", metadata.Header().Get("Cache-Control"))
	var resp blobGrantsResponse
	require.NoError(t, json.Unmarshal(metadata.Body.Bytes(), &resp))
	require.Equal(t, "sha256:"+digest, resp.OCIIndexDigest)
}
