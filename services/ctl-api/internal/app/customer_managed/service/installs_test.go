package service

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func customerManagedTestContext(orgID string) *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	cctx.SetOrgIDGinContext(ctx, orgID)
	ctx.Set(keys.AccountIDCtxKey, strings.Repeat("c", 26))
	return ctx
}

func createCustomerManagedInstallTables(t *testing.T, svc *service) {
	t.Helper()
	require.NoError(t, svc.db.Exec(`CREATE TABLE app_releases (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime, org_id text, app_id text,
		app_config_id text, sandbox_build_id text, component_build_ids json, runbooks json, runtime json, runtime_digest text, definitions_blob text, schema_version integer,
		semantic_digest text, status text, status_description text
	)`).Error)
	require.NoError(t, svc.db.Exec(`CREATE TABLE release_packages (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime, org_id text, release_id text,
		format text, target_platform text, package_digest text, schema_version integer, manifest_digest text, plan_digest text,
		oci_root_digest text, oci_index_digest text, archive_checksum text, archive_size integer, status text, status_description text
	)`).Error)
	require.NoError(t, svc.db.Exec(`CREATE TABLE installs (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime, deleted_at integer default 0,
		metadata text, lifecycle_phase text, labels text, org_id text, name text, app_id text, sandbox_mode boolean,
		app_config_id text, app_branch_id text, app_sandbox_config_id text, app_runner_config_id text,
		component_health_context text, health_cluster_error text, sandbox_health_status text, sandbox_health_message text,
		last_health_report_at datetime, cloud_platform_metadata text, phone_home_auth text
	)`).Error)
	require.NoError(t, svc.db.Exec(`CREATE TABLE install_management_policy_versions (
		id text primary key, created_by_id text, created_at datetime, effective_at datetime, superseded_at datetime,
		org_id text, install_id text, version integer, connectivity text, release_selection text,
		command_authority text, approval_authority text, telemetry text, unique(install_id, version)
	)`).Error)
	require.NoError(t, svc.db.Exec(`CREATE TABLE install_registrations (
		id text primary key, created_by_id text, created_at datetime, org_id text, install_id text,
		release_id text, package_id text, source text, registration text, integrity_status text,
		association_status text, imported_at datetime
	)`).Error)
	require.NoError(t, svc.db.Exec(`CREATE TABLE install_release_deployments (
		id text primary key, created_at datetime, org_id text, install_id text, release_id text, package_id text,
		previous_release_id text, install_app_config_version_id text, policy_version_id text, method text, actor text, executor text, operation_id text, plan_digest text,
		result_directive text, status text, started_at datetime, finished_at datetime
	)`).Error)
}

func seedReleasePackage(t *testing.T, svc *service, orgID string) app.ReleasePackage {
	t.Helper()
	release := app.AppRelease{
		ID: strings.Repeat("r", 26), CreatedByID: strings.Repeat("c", 26), OrgID: orgID,
		AppID: "app-a", AppConfigID: strings.Repeat("f", 26), SemanticDigest: "sha256:" + strings.Repeat("c", 64),
		Status: app.AppReleaseStatusReady, StatusDescription: "ready",
	}
	require.NoError(t, svc.db.Omit("CreatedBy", "Org", "App", "AppConfig", "Members", "Packages").Create(&release).Error)
	pkg := app.ReleasePackage{
		ID: strings.Repeat("p", 26), CreatedByID: strings.Repeat("c", 26), OrgID: orgID, ReleaseID: release.ID,
		Format: app.ReleasePackageFormatPortableOCI, TargetPlatform: "linux/amd64",
		PackageDigest: "sha256:" + strings.Repeat("d", 64), PlanDigest: "sha256:" + strings.Repeat("a", 64),
		ArchiveChecksum: strings.Repeat("b", 64), Status: app.ReleasePackageStatusActive, StatusDescription: "active",
	}
	require.NoError(t, svc.db.Omit("CreatedBy", "Org", "Release", "Members", "Replicas").Create(&pkg).Error)
	return pkg
}

func customerManagedRegistration(t *testing.T, pkg app.ReleasePackage) customermanaged.InstallationRegistration {
	t.Helper()
	registration, err := customermanaged.NewInstallationRegistration(customermanaged.InstallationRegistration{
		SchemaVersion: customermanaged.InstallationRegistrationSchemaVersion,
		ReleaseID:     pkg.ReleaseID, ReleaseDigest: "sha256:" + strings.Repeat("c", 64),
		PackageID: pkg.ID, PackageDigest: pkg.PackageDigest,
		BundleDigest: pkg.PlanDigest, ArchiveDigest: "sha256:" + strings.Repeat("b", 64),
		OperationID:  "install-run-1",
		DeploymentID: "prod", InstallID: "vinst1234-prod",
		Cloud:       customermanaged.InstallationRegistrationCloud{Provider: "aws", AccountID: "123456789012", Region: "us-east-1"},
		Stack:       customermanaged.InstallationRegistrationStack{Type: "aws-cloudformation", ID: "stack-id", Name: "install-stack"},
		InstalledAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	return registration
}

func TestRegisterInstallAtomicallyCreatesDisconnectedRecordsAndIsIdempotent(t *testing.T) {
	svc, _ := testService(t)
	createCustomerManagedInstallTables(t, svc)
	pkg := seedReleasePackage(t, svc, "org-a")
	registration := customerManagedRegistration(t, pkg)
	ctx := customerManagedTestContext("org-a")

	var before int64
	require.NoError(t, svc.db.Model(&app.Install{}).Count(&before).Error)
	require.Zero(t, before)

	first, created, err := svc.registerInstall(ctx, "org-a", registration)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, app.InstallConnectivityDisconnected, first.Policy.Connectivity)
	require.Equal(t, app.InstallAuthorityCustomer, first.Policy.CommandAuthority)
	require.Equal(t, registration.RegistrationID, first.Registration.ID)
	require.Equal(t, pkg.ReleaseID, first.Deployment.ReleaseID)
	require.Equal(t, app.InstallDeploymentMethodDisconnectedLocal, first.Deployment.Method)

	second, created, err := svc.registerInstall(ctx, "org-a", registration)
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, first.Install.ID, second.Install.ID)

	for model, expected := range map[any]int64{
		&app.Install{}: 1, &app.InstallRegistration{}: 1,
		&app.InstallManagementPolicyVersion{}: 1, &app.InstallReleaseDeployment{}: 1,
	} {
		var count int64
		require.NoError(t, svc.db.Model(model).Count(&count).Error)
		require.Equal(t, expected, count)
	}
}
