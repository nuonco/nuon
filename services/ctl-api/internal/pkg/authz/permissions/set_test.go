package permissions

import (
	"encoding/json"
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

	t.Run("distinct verbs from separate policies accumulate", func(t *testing.T) {
		set := Set(NewSet())
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionRead))}))
		require.NoError(t, set.Add(map[string]*string{org: strPtr(string(PermissionCreate))}))

		assert.NoError(t, set.CanPerform(org, PermissionRead))
		assert.NoError(t, set.CanPerform(org, PermissionCreate))
		assert.Error(t, set.CanPerform(org, PermissionUpdate))
		assert.Error(t, set.CanPerform(org, PermissionDelete))
	})
}

func TestGrantVerbSets(t *testing.T) {
	const obj = "app_abc"

	t.Run("verb subset allows only its verbs", func(t *testing.T) {
		set := NewSet()
		set.Grant(obj, PermissionRead, PermissionCreate)

		assert.NoError(t, set.CanPerform(obj, PermissionRead))
		assert.NoError(t, set.CanPerform(obj, PermissionCreate))
		assert.Error(t, set.CanPerform(obj, PermissionUpdate))
		assert.Error(t, set.CanPerform(obj, PermissionDelete))
	})

	t.Run("all collapses and subsumes later grants", func(t *testing.T) {
		set := NewSet()
		set.Grant(obj, PermissionAll)
		set.Grant(obj, PermissionRead)

		assert.Equal(t, Verbs{PermissionAll}, set[obj])
		assert.NoError(t, set.CanPerform(obj, PermissionDelete))
	})

	t.Run("all granted after a subset collapses", func(t *testing.T) {
		set := NewSet()
		set.Grant(obj, PermissionRead)
		set.Grant(obj, PermissionAll)

		assert.Equal(t, Verbs{PermissionAll}, set[obj])
	})

	t.Run("grants merge with hstore adds", func(t *testing.T) {
		set := NewSet()
		require.NoError(t, set.Add(map[string]*string{obj: strPtr(string(PermissionRead))}))
		set.Grant(obj, PermissionUpdate)

		assert.NoError(t, set.CanPerform(obj, PermissionRead))
		assert.NoError(t, set.CanPerform(obj, PermissionUpdate))
		assert.Error(t, set.CanPerform(obj, PermissionDelete))
	})
}

// The account payload's permissions field predates verb sets, and
// already-released clients unmarshal it into map[string]string. The wire shape
// must stay exactly that.
func TestSetMarshalsToLegacyWireShape(t *testing.T) {
	const org = "org_abc"

	cases := []struct {
		name  string
		verbs []Permission
		want  string
	}{
		{name: "all", verbs: []Permission{PermissionAll}, want: "all"},
		{name: "read", verbs: []Permission{PermissionRead}, want: "read"},
		{name: "delete alone", verbs: []Permission{PermissionDelete}, want: "delete"},
		{
			name:  "every verb collapses to all",
			verbs: []Permission{PermissionRead, PermissionCreate, PermissionUpdate, PermissionDelete},
			want:  "all",
		},
		{
			name:  "subset reports its weakest verb rather than overstating",
			verbs: []Permission{PermissionCreate, PermissionRead},
			want:  "read",
		},
		{
			name:  "subset without read reports create",
			verbs: []Permission{PermissionUpdate, PermissionCreate},
			want:  "create",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := NewSet()
			set.Grant(org, tc.verbs...)

			raw, err := json.Marshal(set)
			require.NoError(t, err)

			var legacy map[string]string
			require.NoError(t, json.Unmarshal(raw, &legacy), "a legacy client must be able to parse this")
			assert.Equal(t, tc.want, legacy[org])
		})
	}
}

func TestSetUnmarshalAcceptsBothShapes(t *testing.T) {
	const org = "org_abc"

	t.Run("legacy single verb", func(t *testing.T) {
		var set Set
		require.NoError(t, json.Unmarshal([]byte(`{"org_abc":"all"}`), &set))
		assert.NoError(t, set.CanPerform(org, PermissionDelete))
	})

	t.Run("verb array", func(t *testing.T) {
		var set Set
		require.NoError(t, json.Unmarshal([]byte(`{"org_abc":["read","create"]}`), &set))
		assert.NoError(t, set.CanPerform(org, PermissionRead))
		assert.NoError(t, set.CanPerform(org, PermissionCreate))
		assert.Error(t, set.CanPerform(org, PermissionDelete))
	})

	t.Run("round trip preserves all and single verbs", func(t *testing.T) {
		set := NewSet()
		set.Grant("org_all", PermissionAll)
		set.Grant("app_read", PermissionRead)

		raw, err := json.Marshal(set)
		require.NoError(t, err)

		var back Set
		require.NoError(t, json.Unmarshal(raw, &back))
		assert.NoError(t, back.CanPerform("org_all", PermissionDelete))
		assert.NoError(t, back.CanPerform("app_read", PermissionRead))
		assert.Error(t, back.CanPerform("app_read", PermissionUpdate))
	})
}

func TestNewVerbs(t *testing.T) {
	t.Run("aliases", func(t *testing.T) {
		verbs, err := NewVerbs([]string{"all"})
		require.NoError(t, err)
		assert.Equal(t, Verbs{PermissionAll}, verbs)

		verbs, err = NewVerbs([]string{"read"})
		require.NoError(t, err)
		assert.Equal(t, Verbs{PermissionRead}, verbs)
	})

	t.Run("subset dedupes", func(t *testing.T) {
		verbs, err := NewVerbs([]string{"read", "create", "read"})
		require.NoError(t, err)
		assert.Len(t, verbs, 2)
	})

	t.Run("all subsumes", func(t *testing.T) {
		verbs, err := NewVerbs([]string{"read", "all"})
		require.NoError(t, err)
		assert.Equal(t, Verbs{PermissionAll}, verbs)
	})

	t.Run("invalid verb rejected", func(t *testing.T) {
		_, err := NewVerbs([]string{"deploy"})
		assert.Error(t, err)
	})

	t.Run("empty rejected", func(t *testing.T) {
		_, err := NewVerbs(nil)
		assert.Error(t, err)
	})
}

func TestCanPerformScopedObjects(t *testing.T) {
	const org = "org_abc"
	const scoped = org + ":component_builds"

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

	t.Run("scoped create grant does not allow org-wide mutation", func(t *testing.T) {
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
