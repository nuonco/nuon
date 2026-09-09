package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAppReleaseOwnsOnePackagePerTargetPlatform(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE app_releases (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime, org_id text, app_id text,
		app_config_id text, sandbox_build_id text, component_build_ids json, runbooks json, runtime json, runtime_digest text, definitions_blob text, schema_version integer,
		semantic_digest text, status text, status_description text
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX idx_app_release_identity ON app_releases (org_id, app_id, semantic_digest)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE release_packages (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime, org_id text, release_id text,
		format text, target_platform text, package_digest text, schema_version integer, manifest_digest text, plan_digest text,
		oci_root_digest text, oci_index_digest text, archive_checksum text, archive_size integer, status text,
		status_description text
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX idx_release_package_identity ON release_packages (release_id, format, target_platform)`).Error)

	release := AppRelease{
		ID: strings.Repeat("r", 26), CreatedByID: strings.Repeat("c", 26), OrgID: "org-a", AppID: "app-a",
		AppConfigID: "config-a", SandboxBuildID: "sandbox-build-a", ComponentBuildIDs: map[string]string{"api": "build-a"},
		Runtime: AppReleaseRuntime{RunnerImageURL: "runner", RunnerImageTag: "v1"}, RuntimeDigest: "sha256:runtime",
		SchemaVersion: 1, SemanticDigest: "sha256:release", Status: AppReleaseStatusReady, StatusDescription: "release ready",
	}
	require.NoError(t, db.Omit("CreatedBy", "Org", "App", "AppConfig", "Members", "Packages").Create(&release).Error)

	packages := []ReleasePackage{
		{ID: strings.Repeat("a", 26), CreatedByID: strings.Repeat("c", 26), OrgID: "org-a", ReleaseID: release.ID, Format: ReleasePackageFormatPortableOCI, TargetPlatform: "linux/amd64", PackageDigest: "sha256:amd64", SchemaVersion: 1, Status: ReleasePackageStatusQueued, StatusDescription: "waiting to publish"},
		{ID: strings.Repeat("b", 26), CreatedByID: strings.Repeat("c", 26), OrgID: "org-a", ReleaseID: release.ID, Format: ReleasePackageFormatPortableOCI, TargetPlatform: "linux/arm64", PackageDigest: "sha256:arm64", SchemaVersion: 1, Status: ReleasePackageStatusQueued, StatusDescription: "waiting to publish"},
	}
	for i := range packages {
		require.NoError(t, db.Omit("CreatedBy", "Org", "Release", "Members", "Replicas").Create(&packages[i]).Error)
	}

	var stored AppRelease
	require.NoError(t, db.Preload("Packages").First(&stored, "id = ?", release.ID).Error)
	require.Equal(t, release.SemanticDigest, stored.SemanticDigest)
	require.Len(t, stored.Packages, 2)
	require.ElementsMatch(t, []string{"linux/amd64", "linux/arm64"}, []string{stored.Packages[0].TargetPlatform, stored.Packages[1].TargetPlatform})

	duplicate := packages[0]
	duplicate.ID = strings.Repeat("d", 26)
	require.Error(t, db.Omit("CreatedBy", "Org", "Release", "Members", "Replicas").Create(&duplicate).Error)
}
