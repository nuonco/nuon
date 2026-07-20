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
