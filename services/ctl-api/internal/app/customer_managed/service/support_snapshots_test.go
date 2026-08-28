package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/supportsnapshot"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func createSupportSnapshotsTable(t *testing.T, svc *service) {
	t.Helper()
	require.NoError(t, svc.db.Exec(`CREATE TABLE install_support_snapshots (
		id text primary key, created_by_id text, created_at datetime, org_id text, install_id text,
		archive_sha256 text, archive_size integer, schema_version integer, captured_at datetime,
		storage_provider text, storage_region text, storage_ref text, storage_version text,
		integrity_status text, association_status text, manifest text, snapshot_blob text,
		unique(org_id, archive_sha256)
	)`).Error)
}

func supportSnapshotArchive(t *testing.T, registration customermanaged.InstallationRegistration) []byte {
	return supportSnapshotArchiveWithHistory(t, registration, nil)
}

func supportSnapshotArchiveWithHistory(t *testing.T, registration customermanaged.InstallationRegistration, history []operation.BundleInfo) []byte {
	t.Helper()
	now := time.Now().UTC()
	snapshot := supportsnapshot.Snapshot{
		SchemaVersion: supportsnapshot.SchemaVersion, CapturedAt: now, Registration: registration,
		IncludeState:  true,
		BundleHistory: history,
		State:         &supportsnapshot.CapturedState{Status: json.RawMessage(`{"status":"healthy"}`), Report: json.RawMessage(`{"resources":1}`)},
		Runs:          []supportsnapshot.Run{{RunID: "run-1", RefKind: "drift", RefName: "api", Status: "finished", StartedAt: now}},
		Logs:          []supportsnapshot.JobLog{{JobID: "job-1", Total: 1, Entries: []supportsnapshot.LogEntry{{Level: "info", Msg: "finished"}}}},
		Collection:    supportsnapshot.CollectionReport{SchemaVersion: supportsnapshot.SchemaVersion, Redaction: "support-v1", Included: []string{"state", "runs", "logs"}},
	}
	var archive bytes.Buffer
	_, err := supportsnapshot.Write(&archive, snapshot, supportsnapshot.Producer{Name: "bundle-portal", Version: "test"})
	require.NoError(t, err)
	return archive.Bytes()
}

func TestSupportSnapshotImportIsImmutableAndIdempotent(t *testing.T) {
	svc, store := testService(t)
	createCustomerManagedInstallTables(t, svc)
	createSupportSnapshotsTable(t, svc)
	require.NoError(t, svc.db.Exec(`INSERT INTO orgs (id, features) VALUES (?, ?)`, "org-a", `{"customer-managed-installs":true}`).Error)
	pkg := seedReleasePackage(t, svc, "org-a")
	registration := customerManagedRegistration(t, pkg)
	registeredInstall, registered, err := svc.registerInstall(customerManagedTestContext("org-a"), "org-a", registration)
	require.NoError(t, err)
	require.True(t, registered)
	install := registeredInstall.Install
	store.publishReplica = transport.Replica{Provider: transport.ProviderAWSS3, Region: "us-east-1", StorageRef: "support/archive", StorageVersion: "v1"}
	body := supportSnapshotArchive(t, registration)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		cctx.SetOrgGinContext(ctx, &app.Org{ID: "org-a"})
		ctx.Set(keys.AccountIDCtxKey, strings.Repeat("c", 26))
	})
	require.NoError(t, svc.RegisterPublicRoutes(router))

	request := httptest.NewRequest(http.MethodPost, "/v1/installs/"+install.ID+"/support-snapshots", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var created supportSnapshotResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
	require.Equal(t, supportSnapshotIntegrityVerified, created.IntegrityStatus)
	require.NotNil(t, created.Snapshot)
	require.Len(t, created.Snapshot.Runs, 1)
	require.Equal(t, body, store.published)

	var snapshotBlobMetadata string
	require.NoError(t, svc.db.Raw("SELECT snapshot_blob FROM install_support_snapshots WHERE id = ?", created.ID).Scan(&snapshotBlobMetadata).Error)
	require.Contains(t, snapshotBlobMetadata, `"content_type":"application/json"`)
	require.NotContains(t, snapshotBlobMetadata, `"run_id":"run-1"`)
	blobSvc := svc.blobSvc.(*fakeBlobStore)
	require.Len(t, blobSvc.blobs, 1)
	for _, contents := range blobSvc.blobs {
		require.Contains(t, string(contents), `"status":"healthy"`)
		require.Contains(t, string(contents), `"run_id":"run-1"`)
		require.Contains(t, string(contents), `"msg":"finished"`)
	}

	request = httptest.NewRequest(http.MethodPost, "/v1/installs/"+install.ID+"/support-snapshots", bytes.NewReader(body))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var existing supportSnapshotResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &existing))
	require.Equal(t, created.ID, existing.ID)
	require.NotNil(t, existing.Snapshot)
	require.Len(t, existing.Snapshot.Logs, 1)
	require.Equal(t, 1, blobSvc.downloadCount)

	request = httptest.NewRequest(http.MethodGet, "/v1/installs/"+install.ID+"/support-snapshots", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var listed []supportSnapshotResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &listed))
	require.Len(t, listed, 1)
	require.Nil(t, listed[0].Snapshot)
	require.Equal(t, 1, blobSvc.downloadCount)

	request = httptest.NewRequest(http.MethodGet, "/v1/installs/"+install.ID+"/support-snapshots/"+created.ID, nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var fetched supportSnapshotResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &fetched))
	require.NotNil(t, fetched.Snapshot)
	require.True(t, fetched.Snapshot.IncludeState)
	require.JSONEq(t, `{"status":"healthy"}`, string(fetched.Snapshot.State.Status))
	require.Len(t, fetched.Snapshot.Runs, 1)
	require.Len(t, fetched.Snapshot.Logs, 1)
	require.Equal(t, 2, blobSvc.downloadCount)

	var count int64
	require.NoError(t, svc.db.Model(&app.InstallSupportSnapshot{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestSupportSnapshotRejectsAnotherInstallRegistration(t *testing.T) {
	installed := customermanaged.InstallationRegistration{RegistrationID: "airreg_expected", InstallID: "expected", ReleaseID: "release", PackageID: "package", BundleDigest: "sha256:a", ArchiveDigest: "sha256:b"}
	err := validateSupportSnapshotAssociation(installed, supportsnapshot.Snapshot{Registration: customermanaged.InstallationRegistration{RegistrationID: "airreg_other", InstallID: "other", ReleaseID: "release", PackageID: "package", BundleDigest: "sha256:a", ArchiveDigest: "sha256:b"}})
	require.ErrorContains(t, err, "does not match")
}

func TestSupportSnapshotSynchronizesReleaseDeploymentHistory(t *testing.T) {
	svc, store := testService(t)
	createCustomerManagedInstallTables(t, svc)
	createSupportSnapshotsTable(t, svc)
	require.NoError(t, svc.db.Exec(`INSERT INTO orgs (id, features) VALUES (?, ?)`, "org-a", `{"customer-managed-installs":true}`).Error)
	initialPackage := seedReleasePackage(t, svc, "org-a")
	registration := customerManagedRegistration(t, initialPackage)
	registered, _, err := svc.registerInstall(customerManagedTestContext("org-a"), "org-a", registration)
	require.NoError(t, err)

	release := app.AppRelease{
		ID: strings.Repeat("n", 26), CreatedByID: strings.Repeat("c", 26), OrgID: "org-a",
		AppID: registered.Install.AppID, AppConfigID: strings.Repeat("g", 26),
		SemanticDigest: "sha256:" + strings.Repeat("e", 64), Status: app.AppReleaseStatusReady, StatusDescription: "ready",
	}
	require.NoError(t, svc.db.Omit("CreatedBy", "Org", "App", "AppConfig", "Members", "Packages").Create(&release).Error)
	pkg := app.ReleasePackage{
		ID: strings.Repeat("q", 26), CreatedByID: strings.Repeat("c", 26), OrgID: "org-a", ReleaseID: release.ID,
		Format: app.ReleasePackageFormatPortableOCI, TargetPlatform: "linux/amd64",
		PackageDigest: "sha256:" + strings.Repeat("f", 64), PlanDigest: "sha256:" + strings.Repeat("1", 64),
		ArchiveChecksum: strings.Repeat("2", 64), Status: app.ReleasePackageStatusActive, StatusDescription: "active",
	}
	require.NoError(t, svc.db.Omit("CreatedBy", "Org", "Release", "Members", "Replicas").Create(&pkg).Error)
	activatedAt := registration.InstalledAt.Add(time.Hour).UTC()
	history := []operation.BundleInfo{{
		BundleDigest: pkg.PlanDigest, ActivatedAt: activatedAt,
		Release: &operation.BundleReleaseIdentity{ID: release.ID, Digest: release.SemanticDigest},
		Package: &operation.BundlePackageIdentity{ID: pkg.ID, Digest: pkg.PackageDigest},
	}}
	store.publishReplica = transport.Replica{Provider: transport.ProviderAWSS3, Region: "us-east-1", StorageRef: "support/archive", StorageVersion: "v1"}
	body := supportSnapshotArchiveWithHistory(t, registration, history)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		cctx.SetOrgGinContext(ctx, &app.Org{ID: "org-a"})
		ctx.Set(keys.AccountIDCtxKey, strings.Repeat("c", 26))
	})
	require.NoError(t, svc.RegisterPublicRoutes(router))

	for _, expectedStatus := range []int{http.StatusCreated, http.StatusOK} {
		request := httptest.NewRequest(http.MethodPost, "/v1/installs/"+registered.Install.ID+"/support-snapshots", bytes.NewReader(body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, expectedStatus, response.Code, response.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/installs/"+registered.Install.ID+"/release-deployments", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var deployments []app.InstallReleaseDeployment
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &deployments))
	require.Len(t, deployments, 2)
	require.Equal(t, release.ID, deployments[0].ReleaseID)
	require.Equal(t, initialPackage.ReleaseID, deployments[0].PreviousReleaseID)
	require.Equal(t, pkg.PlanDigest, deployments[0].PlanDigest)
	require.Equal(t, activatedAt, deployments[0].FinishedAt.UTC())
}
