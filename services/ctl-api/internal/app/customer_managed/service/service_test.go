package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedapp "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type fakeBlobStore struct {
	blobs         map[string][]byte
	downloadCount int
}

var _ blobstore.Service = (*fakeBlobStore)(nil)

func (s *fakeBlobStore) Upload(_ context.Context, key string, data []byte) error {
	s.blobs[key] = append([]byte(nil), data...)
	return nil
}

func (s *fakeBlobStore) Delete(_ context.Context, key string) error {
	delete(s.blobs, key)
	return nil
}

func (s *fakeBlobStore) Download(_ context.Context, key string) ([]byte, error) {
	data, ok := s.blobs[key]
	if !ok {
		return nil, fmt.Errorf("blob %s not found", key)
	}
	return append([]byte(nil), data...), nil
}

func (s *fakeBlobStore) UploadStream(ctx context.Context, key string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if err := s.Upload(ctx, key, data); err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func (s *fakeBlobStore) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	data, err := s.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	s.downloadCount++
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *fakeBlobStore) GetMetadata(_ context.Context, key string) (int64, string, error) {
	data, ok := s.blobs[key]
	if !ok {
		return 0, "", fmt.Errorf("blob %s not found", key)
	}
	return int64(len(data)), "application/json", nil
}

type fakeStore struct {
	replica        transport.Replica
	publishReplica transport.Replica
	published      []byte
	filename       string
	grant          transport.DownloadGrant
	blobGrants     map[string]transport.BlobGrant
	granted        []string
}

func (*fakeStore) Configured() bool { return true }

func (s *fakeStore) Publish(_ context.Context, request transport.PublishRequest) (transport.Replica, error) {
	data, err := io.ReadAll(request.Body)
	if err != nil {
		return transport.Replica{}, err
	}
	s.published = data
	return s.publishReplica, nil
}

func (*fakeStore) Delete(context.Context, transport.Replica) error { return nil }

func (s *fakeStore) Grant(_ context.Context, replica transport.Replica, filename string, _ time.Time) (transport.DownloadGrant, error) {
	s.replica = replica
	s.filename = filename
	return s.grant, nil
}

func (*fakeStore) PublishBlob(context.Context, string, string, []byte) error { return nil }

func (s *fakeStore) GrantBlob(_ context.Context, _ string, sha256Hex string) (transport.BlobGrant, error) {
	s.granted = append(s.granted, sha256Hex)
	grant, ok := s.blobGrants[sha256Hex]
	if !ok {
		return transport.BlobGrant{}, fmt.Errorf("blob %s is not available", sha256Hex)
	}
	return grant, nil
}

func testService(t *testing.T) (*service, *fakeStore) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, database.Exec(`CREATE TABLE customer_managed_bundles (
		id text primary key, created_by_id text, created_at datetime, org_id text, app_id text,
		app_config_id text, sandbox_build_id text, component_build_ids json, runbooks json, runbooks_digest text, target_platform text, schema_version integer, manifest_digest text,
		oci_root_digest text, oci_index_digest text, transport_checksum text, size integer, status text, status_description text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE customer_managed_bundle_artifacts (
		id text primary key, org_id text, bundle_id text, kind text, logical_name text,
		component_config_connection_id text, component_id text, action_workflow_id text, app_sandbox_config_id text, config_digest text,
		source_type text, source_identity json, repository text, digest text, media_type text,
		size integer, platform_os text, platform_architecture text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE customer_managed_bundle_transport_replicas (
		id text primary key, created_at datetime, org_id text, bundle_id text, provider text,
		region text, storage_ref text, storage_version text, transport_checksum text, size integer,
		verified_at datetime
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE app_sandbox_builds (
		id text primary key, created_at datetime, deleted_at integer, org_id text, app_id text,
		app_config_id text, app_sandbox_config_id text, status text, status_v2 json
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE orgs (id text primary key, deleted_at integer default 0, features json)`).Error)
	store := &fakeStore{grant: transport.DownloadGrant{URL: "https://download.invalid/signed", ExpiresAt: time.Now().Add(time.Minute), SupportsRange: true}}
	return &service{db: database, store: store, blobSvc: &fakeBlobStore{blobs: make(map[string][]byte)}, features: features.New(features.Params{DB: database})}, store
}

func seedBundle(t *testing.T, svc *service, id, orgID, appID, digest string, createdAt time.Time) app.CustomerManagedBundle {
	t.Helper()
	bundle := app.CustomerManagedBundle{
		ID: id, CreatedByID: strings.Repeat("c", 26), CreatedAt: createdAt, OrgID: orgID, AppID: appID,
		AppConfigID: strings.Repeat("f", 26), TargetPlatform: "linux/amd64", SchemaVersion: 1,
		ManifestDigest: digest, OCIRootDigest: digest, TransportChecksum: digest, Size: 42,
		Status: app.CustomerManagedBundleStatusActive, StatusDescription: "bundle published and verified",
	}
	require.NoError(t, svc.db.Omit("CreatedBy", "Org", "App", "AppConfig", "Artifacts", "Replicas").Create(&bundle).Error)
	return bundle
}

func TestListAndGetIsolateTenantAndExcludeReplicas(t *testing.T) {
	svc, _ := testService(t)
	now := time.Now()
	digest := strings.Repeat("a", 64)
	wanted := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)
	seedBundle(t, svc, strings.Repeat("2", 26), "org-a", "app-a", digest, now.Add(-time.Hour))
	seedBundle(t, svc, strings.Repeat("3", 26), "org-b", "app-a", digest, now.Add(time.Hour))
	seedBundle(t, svc, strings.Repeat("4", 26), "org-a", "app-b", digest, now.Add(time.Hour))
	require.NoError(t, svc.db.Create(&app.CustomerManagedBundleArtifact{ID: strings.Repeat("5", 26), OrgID: "org-a", BundleID: wanted.ID, Kind: "image", LogicalName: "api", Digest: digest, MediaType: "application/test", Size: 4}).Error)
	require.NoError(t, svc.db.Create(&app.CustomerManagedBundleTransportReplica{ID: strings.Repeat("6", 26), CreatedAt: now, OrgID: "org-a", BundleID: wanted.ID, Provider: "secret-provider", StorageRef: "secret-ref", StorageVersion: "v1", TransportChecksum: digest, Size: 42, VerifiedAt: &now}).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	listed, err := svc.listBundles(ctx, "org-a", "app-a")
	require.NoError(t, err)
	require.Len(t, listed, 2)
	require.Equal(t, wanted.ID, listed[0].ID)
	require.Empty(t, listed[0].Artifacts)
	encoded, err := json.Marshal(listed)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "secret-ref")
	require.NotContains(t, string(encoded), "replicas")

	got, err := svc.getBundle(context.Background(), "org-a", "app-a", wanted.ID)
	require.NoError(t, err)
	require.Len(t, got.Artifacts, 1)
	encoded, err = json.Marshal(responseFromBundle(*got))
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "secret-ref")
	_, err = svc.getBundle(context.Background(), "org-b", "app-a", wanted.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = svc.getBundle(context.Background(), "org-a", "app-b", wanted.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateDownloadGrantSelectsNewestVerifiedReplica(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, store := testService(t)
	now := time.Now()
	digest := strings.Repeat("A", 64)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)
	oldVerified := now.Add(-time.Hour)
	newVerified := now
	for _, replica := range []app.CustomerManagedBundleTransportReplica{
		{ID: strings.Repeat("2", 26), CreatedAt: now.Add(time.Hour), OrgID: "org-a", BundleID: bundle.ID, Provider: "unverified", StorageRef: "unverified", StorageVersion: "v", TransportChecksum: digest, Size: 1},
		{ID: strings.Repeat("3", 26), CreatedAt: now, OrgID: "org-a", BundleID: bundle.ID, Provider: "old", StorageRef: "old", StorageVersion: "v", TransportChecksum: digest, Size: 2, VerifiedAt: &oldVerified},
		{ID: strings.Repeat("4", 26), CreatedAt: now, OrgID: "org-a", BundleID: bundle.ID, Provider: "new", StorageRef: "new", StorageVersion: "v", TransportChecksum: digest, Size: 42, VerifiedAt: &newVerified},
	} {
		require.NoError(t, svc.db.Create(&replica).Error)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/", nil)
	ctx.Params = gin.Params{{Key: "app_id", Value: "app-a"}, {Key: "bundle_id", Value: bundle.ID}}
	cctx.SetOrgGinContext(ctx, &app.Org{ID: "org-a"})
	svc.CreateDownloadGrant(ctx)

	require.Equal(t, 200, recorder.Code)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.Equal(t, "new", store.replica.StorageRef)
	require.Equal(t, "app-bundle-aaaaaaaaaaaa.oci.tar.zst", store.filename)
	var response downloadGrantResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "sha256:"+strings.Repeat("a", 64), response.ManifestDigest)
	require.Equal(t, "sha256:"+strings.Repeat("a", 64), response.TransportChecksum)
	require.Equal(t, int64(42), response.Size)
	require.True(t, response.SupportsRange)
}

func TestCreateDownloadGrantFailsWithoutVerifiedReplica(t *testing.T) {
	svc, _ := testService(t)
	now := time.Now()
	digest := strings.Repeat("a", 64)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)
	require.NoError(t, svc.db.Create(&app.CustomerManagedBundleTransportReplica{ID: strings.Repeat("2", 26), CreatedAt: now, OrgID: "org-a", BundleID: bundle.ID, Provider: "test", StorageRef: "ref", StorageVersion: "v", TransportChecksum: digest, Size: 42}).Error)
	_, err := svc.createDownloadGrant(context.Background(), "org-a", "app-a", bundle.ID)
	require.ErrorContains(t, err, "not active or does not have a verified replica")
}

func TestMalformedManifestDigestIsDataError(t *testing.T) {
	svc, _ := testService(t)
	now := time.Now()
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", "bad", now)
	digest := strings.Repeat("a", 64)
	require.NoError(t, svc.db.Create(&app.CustomerManagedBundleTransportReplica{ID: strings.Repeat("2", 26), CreatedAt: now, OrgID: "org-a", BundleID: bundle.ID, Provider: "test", StorageRef: "ref", StorageVersion: "v", TransportChecksum: digest, Size: 42, VerifiedAt: &now}).Error)
	_, err := svc.createDownloadGrant(context.Background(), "org-a", "app-a", bundle.ID)
	require.ErrorContains(t, err, "invalid stored manifest digest")
}

func TestResolveActiveBuildsMatchesExactAppConfig(t *testing.T) {
	svc, _ := testService(t)
	createReleaseBuildTables(t, svc.db)
	cfg := &app.AppConfig{ID: "config-a", SandboxConfig: app.AppSandboxConfig{ID: "sandbox-a"}}

	_, err := svc.resolveActiveBuilds(context.Background(), "org-a", "app-a", cfg)
	var precondition preconditionError
	require.ErrorAs(t, err, &precondition)
	require.ErrorContains(t, err, "no active sandbox build for app config config-a")

	builds := []app.AppSandboxBuild{
		{ID: strings.Repeat("1", 26), CreatedAt: time.Now(), OrgID: "org-b", AppID: "app-a", AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive},
		{ID: strings.Repeat("2", 26), CreatedAt: time.Now(), OrgID: "org-a", AppID: "app-b", AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive},
		{ID: strings.Repeat("3", 26), CreatedAt: time.Now(), OrgID: "org-a", AppID: "app-a", AppConfigID: "config-b", AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive},
		{ID: strings.Repeat("4", 26), CreatedAt: time.Now(), OrgID: "org-a", AppID: "app-a", AppConfigID: cfg.ID, AppSandboxConfigID: "sandbox-b", Status: app.AppSandboxBuildStatusActive},
		{ID: strings.Repeat("5", 26), CreatedAt: time.Now(), OrgID: "org-a", AppID: "app-a", AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusError},
	}
	for i := range builds {
		require.NoError(t, svc.db.Exec(`INSERT INTO app_sandbox_builds
			(id, created_at, deleted_at, org_id, app_id, app_config_id, app_sandbox_config_id, status)
			VALUES (?, ?, 0, ?, ?, ?, ?, ?)`, builds[i].ID, builds[i].CreatedAt, builds[i].OrgID, builds[i].AppID, builds[i].AppConfigID, builds[i].AppSandboxConfigID, builds[i].Status).Error)
	}
	_, err = svc.resolveActiveBuilds(context.Background(), "org-a", "app-a", cfg)
	require.ErrorAs(t, err, &precondition)

	active := app.AppSandboxBuild{ID: strings.Repeat("6", 26), CreatedAt: time.Now(), OrgID: "org-a", AppID: "app-a", AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive}
	require.NoError(t, svc.db.Exec(`INSERT INTO app_sandbox_builds
		(id, created_at, deleted_at, org_id, app_id, app_config_id, app_sandbox_config_id, status)
		VALUES (?, ?, 0, ?, ?, ?, ?, ?)`, active.ID, active.CreatedAt, active.OrgID, active.AppID, active.AppConfigID, active.AppSandboxConfigID, active.Status).Error)
	selection, err := svc.resolveActiveBuilds(context.Background(), "org-a", "app-a", cfg)
	require.NoError(t, err)
	require.Equal(t, active.ID, selection.sandboxBuildID)
	require.Empty(t, selection.componentBuildIDs)
}

func TestResolveActiveBuildsReusesReleasedEquivalentSandbox(t *testing.T) {
	svc, _ := testService(t)
	createReleaseBuildTables(t, svc.db)
	cfg := &app.AppConfig{ID: "config-new", SandboxConfig: app.AppSandboxConfig{ID: "sandbox-new", Type: "aws-eks"}}
	build := app.AppSandboxBuild{ID: strings.Repeat("1", 26), CreatedAt: time.Now(), OrgID: "org-a", AppID: "app-a", AppConfigID: "config-old", AppSandboxConfigID: "sandbox-old", Status: app.AppSandboxBuildStatusActive}
	require.NoError(t, svc.db.Exec(`INSERT INTO app_sandbox_builds
		(id, created_at, deleted_at, org_id, app_id, app_config_id, app_sandbox_config_id, status)
		VALUES (?, ?, 0, ?, ?, ?, ?, ?)`, build.ID, build.CreatedAt, build.OrgID, build.AppID, build.AppConfigID, build.AppSandboxConfigID, build.Status).Error)
	require.NoError(t, svc.db.Exec(`INSERT INTO app_releases (id, created_at, org_id, app_id, status) VALUES (?, ?, ?, ?, ?)`,
		strings.Repeat("2", 26), time.Now(), "org-a", "app-a", app.AppReleaseStatusReady).Error)
	definition, err := customermanagedapp.CanonicalObject(cfg.SandboxConfig)
	require.NoError(t, err)
	require.NoError(t, svc.db.Exec(`INSERT INTO app_release_members (id, org_id, release_id, kind, config_digest, build_id) VALUES (?, ?, ?, ?, ?, ?)`,
		strings.Repeat("3", 26), "org-a", strings.Repeat("2", 26), "sandbox", customermanagedapp.ObjectDigest(definition), build.ID).Error)

	selection, err := svc.resolveActiveBuilds(context.Background(), "org-a", "app-a", cfg)
	require.NoError(t, err)
	require.Equal(t, build.ID, selection.sandboxBuildID)
}

func createReleaseBuildTables(t *testing.T, database *gorm.DB) {
	t.Helper()
	require.NoError(t, database.Exec(`CREATE TABLE app_releases (
		id text primary key, created_at datetime, org_id text, app_id text, status text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE app_release_members (
		id text primary key, org_id text, release_id text, kind text, config_digest text, build_id text
	)`).Error)
}
