package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedapp "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
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
	require.NoError(t, database.Exec(`CREATE TABLE app_sandbox_builds (
		id text primary key, created_at datetime, deleted_at integer, org_id text, app_id text,
		app_config_id text, app_sandbox_config_id text, status text, status_v2 json
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE orgs (id text primary key, deleted_at integer default 0, features json)`).Error)
	store := &fakeStore{grant: transport.DownloadGrant{URL: "https://download.invalid/signed", ExpiresAt: time.Now().Add(time.Minute), SupportsRange: true}}
	return &service{db: database, store: store, blobSvc: &fakeBlobStore{blobs: make(map[string][]byte)}, features: features.New(features.Params{DB: database})}, store
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

func TestResolveActiveBuildsUsesPinnedBuildFromEquivalentConnection(t *testing.T) {
	svc, _ := testService(t)
	createReleaseBuildTables(t, svc.db)
	buildID := strings.Repeat("1", 26)
	originalConnectionID := strings.Repeat("2", 26)
	targetConnectionID := strings.Repeat("3", 26)
	componentID := strings.Repeat("4", 26)
	require.NoError(t, svc.db.Exec(`INSERT INTO component_config_connections (id, org_id, component_id) VALUES (?, ?, ?)`, originalConnectionID, "org-a", componentID).Error)
	require.NoError(t, svc.db.Exec(`INSERT INTO component_builds (id, created_at, deleted_at, org_id, component_config_connection_id, status) VALUES (?, ?, 0, ?, ?, ?)`, buildID, time.Now(), "org-a", originalConnectionID, app.ComponentBuildStatusActive).Error)
	require.NoError(t, svc.db.Exec(`INSERT INTO app_sandbox_builds (id, created_at, deleted_at, org_id, app_id, app_config_id, app_sandbox_config_id, status) VALUES (?, ?, 0, ?, ?, ?, ?, ?)`, strings.Repeat("5", 26), time.Now(), "org-a", "app-a", "config-a", "sandbox-a", app.AppSandboxBuildStatusActive).Error)

	cfg := &app.AppConfig{
		ID: "config-a",
		ComponentConfigConnections: []app.ComponentConfigConnection{{
			ID: targetConnectionID, ComponentID: componentID, ComponentName: "component-a",
			LatestBuildID: generics.NullString{NullString: sql.NullString{String: buildID, Valid: true}},
		}},
		SandboxConfig: app.AppSandboxConfig{ID: "sandbox-a"},
	}
	selection, err := svc.resolveActiveBuilds(context.Background(), "org-a", "app-a", cfg)
	require.NoError(t, err)
	require.Equal(t, buildID, selection.componentBuildIDs[targetConnectionID])
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
	require.NoError(t, database.Exec(`CREATE TABLE component_config_connections (
		id text primary key, deleted_at integer default 0, org_id text, component_id text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE component_builds (
		id text primary key, created_at datetime, deleted_at integer, org_id text,
		component_config_connection_id text, status text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE app_releases (
		id text primary key, created_at datetime, org_id text, app_id text, status text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE app_release_members (
		id text primary key, org_id text, release_id text, kind text, config_digest text, build_id text
	)`).Error)
}
