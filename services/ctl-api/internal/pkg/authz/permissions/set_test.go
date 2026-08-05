package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestSetAddKeepsStrongerGrant(t *testing.T) {
	const org = "org_abc"

	t.Run("all is not clobbered by a later read", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionAll))}))
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionRead))}))

		assert.NoError(t, set.CanPerform(org, PermissionDelete))
	})

	t.Run("read is upgraded to all", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionRead))}))
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionAll))}))

		assert.NoError(t, set.CanPerform(org, PermissionDelete))
	})

	t.Run("read-only alone cannot mutate", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionRead))}))

		assert.NoError(t, set.CanPerform(org, PermissionRead))
		assert.Error(t, set.CanPerform(org, PermissionCreate))
		assert.Error(t, set.CanPerform(org, PermissionDelete))
	})
}

func TestCanPerformScopedObjects(t *testing.T) {
	const org = "org_abc"
	scoped := ComponentBuildsObject(org)

	t.Run("org-wide all grant covers scoped objects", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionAll))}))

		assert.NoError(t, set.CanPerform(scoped, PermissionCreate))
	})

	t.Run("org-wide read grant does not allow scoped create", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionRead))}))

		assert.NoError(t, set.CanPerform(scoped, PermissionRead))
		assert.Error(t, set.CanPerform(scoped, PermissionCreate))
	})

	t.Run("builder grant allows scoped create but not org-wide mutation", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{
			org:    strPtr(string(PermissionRead)),
			scoped: strPtr(string(PermissionCreate)),
		}))

		assert.NoError(t, set.CanPerform(scoped, PermissionCreate))
		assert.NoError(t, set.CanPerform(org, PermissionRead))
		assert.Error(t, set.CanPerform(scoped, PermissionDelete))
		assert.Error(t, set.CanPerform(org, PermissionCreate))
	})

	t.Run("org-wide all remains additive with a scoped grant", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{
			org:    strPtr(string(PermissionAll)),
			scoped: strPtr(string(PermissionCreate)),
		}))

		assert.NoError(t, set.CanPerform(scoped, PermissionCreate))
		assert.NoError(t, set.CanPerform(scoped, PermissionDelete))
	})

	t.Run("wildcard all remains additive with a scoped grant", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{
			"*":    strPtr(string(PermissionAll)),
			scoped: strPtr(string(PermissionCreate)),
		}))

		assert.NoError(t, set.CanPerform(scoped, PermissionDelete))
	})

	t.Run("no grants at all is denied", func(t *testing.T) {
		set := Set(NewSet())

		assert.Error(t, set.CanPerform(scoped, PermissionCreate))
	})
}
