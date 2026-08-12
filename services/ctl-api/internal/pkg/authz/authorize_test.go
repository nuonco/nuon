package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestAuthorize(t *testing.T) {
	const (
		orgID    = "orgxxxxxxxxxxxxxxxxxxxxxxx"
		appWeb   = "appwebxxxxxxxxxxxxxxxxxxxx"
		appBill  = "appbillxxxxxxxxxxxxxxxxxxx"
		brnWeb   = "brnwebxxxxxxxxxxxxxxxxxxxx"
		brnBill  = "brnbillxxxxxxxxxxxxxxxxxxx"
		instProd = "instprodxxxxxxxxxxxxxxxxxx"
	)

	branchChain := func(appID, branchID string) []Link {
		return []Link{
			{Type: app.LevelAppBranch, ID: branchID},
			{Type: app.LevelApp, ID: appID},
			{Type: app.LevelOrg, ID: orgID},
		}
	}
	installChain := []Link{
		{Type: app.LevelInstall, ID: instProd},
		{Type: app.LevelApp, ID: appWeb},
		{Type: app.LevelOrg, ID: orgID},
	}

	t.Run("object grant on the leaf authorizes", func(t *testing.T) {
		perms := permissions.NewSet()
		perms.Grant(instProd, permissions.PermissionRead)

		assert.NoError(t, Authorize(perms, nil, installChain, permissions.PermissionRead))
		assert.Error(t, Authorize(perms, nil, installChain, permissions.PermissionDelete))
	})

	t.Run("object grant on an ancestor flows down", func(t *testing.T) {
		perms := permissions.NewSet()
		perms.Grant(appWeb, permissions.PermissionRead)

		assert.NoError(t, Authorize(perms, nil, installChain, permissions.PermissionRead))
		assert.NoError(t, Authorize(perms, nil, branchChain(appWeb, brnWeb), permissions.PermissionRead))
		assert.Error(t, Authorize(perms, nil, branchChain(appBill, brnBill), permissions.PermissionRead))
	})

	t.Run("org-tier grant authorizes everything under the org", func(t *testing.T) {
		perms := permissions.NewSet()
		perms.Grant(orgID, permissions.PermissionAll)

		assert.NoError(t, Authorize(perms, nil, installChain, permissions.PermissionDelete))
		assert.NoError(t, Authorize(perms, nil, branchChain(appBill, brnBill), permissions.PermissionCreate))
	})

	t.Run("org-wide type wildcard authorizes across apps", func(t *testing.T) {
		wildcards := map[app.Level][]app.TypeGrant{
			app.LevelAppBranch: {{Verbs: permissions.Verbs{permissions.PermissionAll}}},
		}

		assert.NoError(t, Authorize(permissions.NewSet(), wildcards, branchChain(appWeb, brnWeb), permissions.PermissionDelete))
		assert.NoError(t, Authorize(permissions.NewSet(), wildcards, branchChain(appBill, brnBill), permissions.PermissionDelete))
		assert.Error(t, Authorize(permissions.NewSet(), wildcards, installChain, permissions.PermissionRead))
	})

	t.Run("scoped wildcard matches only under its parent", func(t *testing.T) {
		wildcards := map[app.Level][]app.TypeGrant{
			app.LevelAppBranch: {{ScopeID: appWeb, Verbs: permissions.Verbs{permissions.PermissionAll}}},
		}

		assert.NoError(t, Authorize(permissions.NewSet(), wildcards, branchChain(appWeb, brnWeb), permissions.PermissionDelete))
		assert.Error(t, Authorize(permissions.NewSet(), wildcards, branchChain(appBill, brnBill), permissions.PermissionDelete))
	})

	t.Run("scoped wildcard verbs still bound the request", func(t *testing.T) {
		wildcards := map[app.Level][]app.TypeGrant{
			app.LevelAppBranch: {{ScopeID: appWeb, Verbs: permissions.Verbs{permissions.PermissionRead}}},
		}

		assert.NoError(t, Authorize(permissions.NewSet(), wildcards, branchChain(appWeb, brnWeb), permissions.PermissionRead))
		assert.Error(t, Authorize(permissions.NewSet(), wildcards, branchChain(appWeb, brnWeb), permissions.PermissionDelete))
	})

	t.Run("scoped wildcard does not authorize the parent itself", func(t *testing.T) {
		wildcards := map[app.Level][]app.TypeGrant{
			app.LevelAppBranch: {{ScopeID: appWeb, Verbs: permissions.Verbs{permissions.PermissionAll}}},
		}
		appChain := []Link{
			{Type: app.LevelApp, ID: appWeb},
			{Type: app.LevelOrg, ID: orgID},
		}

		assert.Error(t, Authorize(permissions.NewSet(), wildcards, appChain, permissions.PermissionRead))
	})

	t.Run("no grants is denied with the leaf id", func(t *testing.T) {
		err := Authorize(permissions.NewSet(), nil, installChain, permissions.PermissionRead)
		assert.ErrorContains(t, err, instProd)
	})
}
