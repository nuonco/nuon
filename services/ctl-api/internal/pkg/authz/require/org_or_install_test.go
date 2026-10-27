package require

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func runOrgOrInstall(t *testing.T, method string, perms permissions.Set, runner *app.Runner) *gin.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	var ctx *gin.Context
	engine.Use(func(c *gin.Context) {
		ctx = c
		if perms != nil {
			cctx.SetAccountGinContext(c, &app.Account{ID: "acct_one", AllPermissions: perms})
		}
		cctx.SetOrgIDGinContext(c, orgID)
		if runner != nil {
			cctx.SetRunnerGinContext(c, runner)
		}
	})
	OrgOrInstallRoute(engine).POST("/v1/terraform-workspaces/:workspace_id/state-json", func(*gin.Context) {})
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, "/v1/terraform-workspaces/ws_one/state-json", nil))
	return ctx
}

func installScopedPerms(t *testing.T) permissions.Set {
	t.Helper()
	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(map[string]*string{
		permissions.Object(orgID, permissions.KindInstall, installID): permissions.PermissionAll.ToStrPtr(),
	}))
	return set
}

func TestOrgOrInstallRoute(t *testing.T) {
	installRunner := &app.Runner{
		ID: "runner_one",
		RunnerGroup: app.RunnerGroup{
			Type:    app.RunnerGroupTypeInstall,
			OwnerID: installID,
		},
	}

	t.Run("org-wide grant passes without runner", func(t *testing.T) {
		set := permissions.Set(permissions.NewSet())
		require.NoError(t, set.Add(map[string]*string{orgID: permissions.PermissionAll.ToStrPtr()}))
		assertAllowed(t, runOrgOrInstall(t, http.MethodPost, set, nil))
	})

	t.Run("install grant with install runner passes", func(t *testing.T) {
		assertAllowed(t, runOrgOrInstall(t, http.MethodPost, installScopedPerms(t), installRunner))
	})

	t.Run("install grant without runner is denied", func(t *testing.T) {
		ctx := runOrgOrInstall(t, http.MethodPost, installScopedPerms(t), nil)
		require.True(t, ctx.IsAborted())
		require.Len(t, ctx.Errors, 1)
		var denied stderr.ErrAuthorization
		require.ErrorAs(t, ctx.Errors[0].Err, &denied)
	})

	t.Run("install grant for other install is denied", func(t *testing.T) {
		set := permissions.Set(permissions.NewSet())
		require.NoError(t, set.Add(map[string]*string{
			permissions.Object(orgID, permissions.KindInstall, otherInstallID): permissions.PermissionAll.ToStrPtr(),
		}))
		ctx := runOrgOrInstall(t, http.MethodPost, set, installRunner)
		require.True(t, ctx.IsAborted())
		assert.NotEmpty(t, ctx.Errors)
	})
}
