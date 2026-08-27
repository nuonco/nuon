package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	assert.Equal(t, "org_one:app/app_one", Object("org_one", KindApp, "app_one"))
	assert.Equal(t, "org_one:install/install_one", Object("org_one", KindInstall, "install_one"))
	assert.Equal(t, "org_one:stack/install_one", Object("org_one", KindStack, "install_one"))

	// StackObject must keep the exact key shape already written to role rows.
	assert.Equal(t, "org_one:stack/install_one", StackObject("org_one", "install_one"))
}

// Only one ":" so CanPerform's parent fallback resolves to the org, not a kind.
func TestObjectParentIsTheOrg(t *testing.T) {
	const org = "org_one"
	obj := Object(org, KindStack, "install_one")

	set := Set(NewSet())
	require.NoError(t, set.Add(map[string]*string{org: PermissionAll.ToStrPtr()}))

	assert.NoError(t, set.CanPerform(obj, PermissionRead))
}
