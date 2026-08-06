package service

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/stretchr/testify/require"
)

// Mirror the auth middleware so the Install BeforeCreate hook can stamp
// org and account IDs.
func airgapTestContext(orgID string) *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	cctx.SetOrgIDGinContext(ctx, orgID)
	ctx.Set(keys.AccountIDCtxKey, strings.Repeat("c", 26))
	return ctx
}

func createInstallsTable(t *testing.T, svc *service) {
	t.Helper()
	require.NoError(t, svc.db.Exec(`CREATE TABLE installs (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime, deleted_at integer default 0,
		metadata text, lifecycle_phase text, labels text, org_id text, name text, app_id text, sandbox_mode boolean,
		app_config_id text, airgap_bundle_id text, app_branch_id text, app_sandbox_config_id text, app_runner_config_id text,
		component_health_context text, health_cluster_error text, sandbox_health_status text, sandbox_health_message text,
		last_health_report_at datetime, cloud_platform_metadata text, phone_home_auth text
	)`).Error)
}

func TestCreateAirgapInstallPersistsOnlyTheInstallRow(t *testing.T) {
	svc, _ := testService(t)
	createInstallsTable(t, svc)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", strings.Repeat("a", 64), time.Now())

	ctx := airgapTestContext("org-a")
	install, err := svc.createAirgapInstall(ctx, "org-a", "app-a", bundle.ID, "acme-prod")
	require.NoError(t, err)
	require.Len(t, install.ID, 26)
	require.True(t, strings.HasPrefix(install.ID, "inl"))
	require.Equal(t, "acme-prod", install.Name)
	require.Equal(t, "app-a", install.AppID)
	require.Equal(t, bundle.AppConfigID, install.AppConfigID)
	require.Equal(t, bundle.ID, install.AirgapBundleID.String)

	var count int64
	require.NoError(t, svc.db.Table("installs").Count(&count).Error)
	require.EqualValues(t, 1, count)

	// The virtual install must not trigger any provisioning: no runner group,
	// queue, or workflow tables exist in this schema, so any attempt to write
	// them would have errored above.
	tables := []string{}
	require.NoError(t, svc.db.Raw(`SELECT name FROM sqlite_master WHERE type='table'`).Scan(&tables).Error)
	for _, table := range tables {
		require.NotContains(t, []string{"runner_groups", "queues", "workflows", "install_sandboxes"}, table)
	}
}

func TestCreateAirgapInstallRejectsForeignBundle(t *testing.T) {
	svc, _ := testService(t)
	createInstallsTable(t, svc)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-b", "app-a", strings.Repeat("a", 64), time.Now())

	ctx := airgapTestContext("org-a")
	_, err := svc.createAirgapInstall(ctx, "org-a", "app-a", bundle.ID, "acme-prod")
	require.Error(t, err)

	_, err = svc.createAirgapInstall(ctx, "org-b", "app-other", bundle.ID, "acme-prod")
	require.Error(t, err)

	var count int64
	require.NoError(t, svc.db.Table("installs").Count(&count).Error)
	require.EqualValues(t, 0, count)
}

func TestListAirgapInstallsIsolatesTenantAndBundle(t *testing.T) {
	svc, _ := testService(t)
	createInstallsTable(t, svc)
	now := time.Now()
	digest := strings.Repeat("a", 64)
	bundle := seedBundle(t, svc, strings.Repeat("1", 26), "org-a", "app-a", digest, now)
	other := seedBundle(t, svc, strings.Repeat("2", 26), "org-a", "app-a", digest, now.Add(-time.Hour))

	ctx := airgapTestContext("org-a")
	first, err := svc.createAirgapInstall(ctx, "org-a", "app-a", bundle.ID, "acme-prod")
	require.NoError(t, err)
	_, err = svc.createAirgapInstall(ctx, "org-a", "app-a", other.ID, "acme-staging")
	require.NoError(t, err)

	listed, err := svc.listAirgapInstalls(ctx, "org-a", "app-a", bundle.ID)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, first.ID, listed[0].ID)
	require.Equal(t, "acme-prod", listed[0].Name)
	require.Equal(t, bundle.ID, listed[0].AirgapBundleID)

	listed, err = svc.listAirgapInstalls(ctx, "org-b", "app-a", bundle.ID)
	require.NoError(t, err)
	require.Empty(t, listed)
}
