package apps

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func testResolver(nameToID map[string]string) *branchNameResolver {
	r := newBranchNameResolver(nil, "app-1")
	for name, id := range nameToID {
		r.installNameToID[name] = id
		r.installIDToName[id] = name
	}
	r.installsLoaded = true
	return r
}

// The API rejects a preview config carrying both install_id and install_name,
// so a named install has to go out as an ID alone.
func TestPreviewConfigRequestInstallNameSendsIDOnly(t *testing.T) {
	resolver := testResolver(map[string]string{"byoc-aws": "inl-1"})

	out, err := previewConfigRequest(context.Background(), resolver, &config.AppBranchConfig{
		Name: "default",
		Preview: &config.AppBranchPreviewConfig{
			Mode:        "plan-only",
			InstallName: "byoc-aws",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "inl-1", out.InstallID)
	require.Empty(t, out.InstallName)
}

func TestPreviewConfigRequestInstallIDPassesThrough(t *testing.T) {
	out, err := previewConfigRequest(context.Background(), testResolver(nil), &config.AppBranchConfig{
		Name: "default",
		Preview: &config.AppBranchPreviewConfig{
			Mode:      "plan-only",
			InstallID: "inl-1",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "inl-1", out.InstallID)
	require.Empty(t, out.InstallName)
}

func TestPreviewConfigRequestUnknownInstallName(t *testing.T) {
	_, err := previewConfigRequest(context.Background(), testResolver(nil), &config.AppBranchConfig{
		Name: "default",
		Preview: &config.AppBranchPreviewConfig{
			Mode:        "plan-only",
			InstallName: "nope",
		},
	})
	require.ErrorContains(t, err, "unknown install name")
}
