package roles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePermissionEntries(t *testing.T) {
	entries, err := ParsePermissionEntries([]string{
		"read:app:app_web",
		"all:install:inl4plkdhwau58atwfd92vlc8q",
		"read,update:app_branch:*:scope=app_web",
	})
	require.NoError(t, err)
	require.Len(t, entries, 3)

	require.Equal(t, "app", *entries[0].ResourceType)
	require.Equal(t, "app_web", *entries[0].ResourceID)
	require.Equal(t, []string{"read"}, entries[0].Permissions)
	require.Empty(t, entries[0].ScopeType)

	require.Equal(t, []string{"all"}, entries[1].Permissions)

	require.Equal(t, []string{"read", "update"}, entries[2].Permissions)
	require.Equal(t, "*", *entries[2].ResourceID)
	require.Equal(t, "app", entries[2].ScopeType)
	require.Equal(t, "app_web", entries[2].ScopeID)
}

func TestParsePermissionEntriesRejects(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  string
		msg  string
	}{
		{"too few segments", "read:app", "expected verbs:resource_type:resource"},
		{"too many segments", "read:app:a:scope=b:c", "expected verbs:resource_type:resource"},
		{"unknown verb", "list:app:app_web", `unknown verb "list"`},
		{"unknown resource type", "read:runner:x", `unknown resource type "runner"`},
		{"empty verbs", ":app:app_web", "at least one verb is required"},
		{"empty resource", "read:app:", "resource is required"},
		{"malformed scope", "read:install:*:app_web", "expected scope=<app id or name>"},
		// The API rejects a scope on a specific resource; catching it here keeps
		// the error next to the flag that caused it.
		{"scope on specific resource", "read:install:inl_abc:scope=app_web", "scope only applies to wildcard"},
		{"scope on app wildcard", "read:app:*:scope=app_web", "app wildcards cannot be scoped"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePermissionEntries([]string{tc.val})
			require.ErrorContains(t, err, tc.msg)
		})
	}
}

func TestParsePermissionEntriesRequiresOne(t *testing.T) {
	_, err := ParsePermissionEntries(nil)
	require.ErrorContains(t, err, "at least one --permission is required")
}
