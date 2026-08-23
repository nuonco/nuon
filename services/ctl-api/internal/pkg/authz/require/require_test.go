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

const (
	orgID          = "org_one"
	installID      = "install_one"
	otherInstallID = "install_two"
)

// run drives the middleware the way the engine does: account and org already
// resolved by the auth and runner_org middlewares upstream of it.
func run(t *testing.T, mw gin.HandlerFunc, method string, perms permissions.Set, orgID, installID string) *gin.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(method, "/v1/stacks/"+installID+"/config", nil)
	ctx.Params = gin.Params{{Key: "install_id", Value: installID}}

	cctx.SetAccountGinContext(ctx, &app.Account{ID: "acct_one", AllPermissions: perms})
	cctx.SetOrgIDGinContext(ctx, orgID)

	mw(ctx)

	return ctx
}

func readStack(t *testing.T, perms permissions.Set, orgID, installID string) *gin.Context {
	t.Helper()

	return run(t, Route(permissions.KindStack, permissions.PermissionRead, "install_id"),
		http.MethodGet, perms, orgID, installID)
}

// A 403 would confirm a resource exists in another org, so the middleware has to
// deny the same way the handlers do.
func assertNotFound(t *testing.T, ctx *gin.Context, description string) {
	t.Helper()

	require.True(t, ctx.IsAborted(), "a denied request must not reach the handler")
	require.Len(t, ctx.Errors, 1)

	var nfErr stderr.ErrNotFound
	require.ErrorAs(t, ctx.Errors[0].Err, &nfErr)
	assert.Equal(t, description, nfErr.Description)

	// The stderr handler serializes Err into the response body, so it must carry
	// nothing beyond the description: a denial reason there would confirm the
	// resource exists.
	assert.Equal(t, description, nfErr.Error(),
		"client-visible error must not leak the denial reason")
}

func assertAllowed(t *testing.T, ctx *gin.Context) {
	t.Helper()

	assert.False(t, ctx.IsAborted())
	assert.Empty(t, ctx.Errors)
}

func scopedStack(t *testing.T, orgID, installID string, perm permissions.Permission) permissions.Set {
	t.Helper()

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(map[string]*string{
		permissions.Object(orgID, permissions.KindStack, installID): perm.ToStrPtr(),
	}))

	return set
}

func TestRoute(t *testing.T) {
	t.Run("scoped grant reaches its own resource", func(t *testing.T) {
		assertAllowed(t, readStack(t, scopedStack(t, orgID, installID, permissions.PermissionRead), orgID, installID))
	})

	// The whole point of the scoped role: one stack's token cannot read another
	// install's config, even inside the same org.
	t.Run("scoped grant cannot reach another resource in the org", func(t *testing.T) {
		ctx := readStack(t, scopedStack(t, orgID, otherInstallID, permissions.PermissionRead), orgID, installID)
		assertNotFound(t, ctx, "install not found")
	})

	// Intentional: the permission set falls back to the object's parent key, so
	// existing org-wide tokens keep working.
	t.Run("org-wide grant passes via parent fallback", func(t *testing.T) {
		set := permissions.Set(permissions.NewSet())
		require.NoError(t, set.Add(map[string]*string{orgID: permissions.PermissionAll.ToStrPtr()}))

		assertAllowed(t, readStack(t, set, orgID, installID))
	})

	t.Run("no grant at all", func(t *testing.T) {
		assertNotFound(t, readStack(t, permissions.Set(permissions.NewSet()), orgID, installID), "install not found")
	})

	// A grant for the same install ID under a different org must not carry over.
	t.Run("scoped grant in another org", func(t *testing.T) {
		assertNotFound(t, readStack(t, scopedStack(t, "org_two", installID, permissions.PermissionRead), orgID, installID), "install not found")
	})

	t.Run("missing account", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/stacks/"+installID+"/config", nil)

		Route(permissions.KindStack, permissions.PermissionRead, "install_id")(ctx)
		assertNotFound(t, ctx, "install not found")
	})

	t.Run("missing org", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/stacks/"+installID+"/config", nil)
		cctx.SetAccountGinContext(ctx, &app.Account{ID: "acct_one", AllPermissions: permissions.Set(permissions.NewSet())})

		Route(permissions.KindStack, permissions.PermissionRead, "install_id")(ctx)
		assertNotFound(t, ctx, "install not found")
	})

	t.Run("missing path param", func(t *testing.T) {
		set := scopedStack(t, orgID, installID, permissions.PermissionRead)
		assertNotFound(t, readStack(t, set, orgID, ""), "install not found")
	})

	t.Run("kind names the resource in the error", func(t *testing.T) {
		ctx := run(t, Route(permissions.KindApp, permissions.PermissionRead, "install_id"),
			http.MethodGet, permissions.Set(permissions.NewSet()), orgID, installID)
		assertNotFound(t, ctx, "app not found")
	})
}

// The declared verb is what gets checked, not the one FromRequest infers from
// the method. Every case here is a POST, which FromRequest would call create.
func TestRouteEnforcesDeclaredVerb(t *testing.T) {
	update := Route(permissions.KindStack, permissions.PermissionUpdate, "install_id")

	t.Run("declared update passes on an update-only grant", func(t *testing.T) {
		set := scopedStack(t, orgID, installID, permissions.PermissionUpdate)
		assertAllowed(t, run(t, update, http.MethodPost, set, orgID, installID))
	})

	// The method's inferred verb must not stand in for the declared one.
	t.Run("declared update denied on a create-only grant", func(t *testing.T) {
		set := scopedStack(t, orgID, installID, permissions.PermissionCreate)
		assertNotFound(t, run(t, update, http.MethodPost, set, orgID, installID), "install not found")
	})

	// And a read-declared route is not widened by a mutating method.
	t.Run("declared read passes on a read-only grant despite a POST", func(t *testing.T) {
		set := scopedStack(t, orgID, installID, permissions.PermissionRead)
		ctx := run(t, Route(permissions.KindStack, permissions.PermissionRead, "install_id"),
			http.MethodPost, set, orgID, installID)
		assertAllowed(t, ctx)
	})
}
