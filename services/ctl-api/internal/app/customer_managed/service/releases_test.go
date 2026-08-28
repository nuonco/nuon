package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	releasearchive "github.com/nuonco/nuon/pkg/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSemanticReleaseDigestIsCanonicalAndPackageIndependent(t *testing.T) {
	members := []app.AppReleaseMember{
		{Kind: "sandbox", LogicalName: "sandbox", BuildID: "build-sandbox", ConfigDigest: "sha256:sandbox-config", ContentDigest: "sha256:sandbox-content"},
		{Kind: "component", LogicalName: "api", BuildID: "build-api", ConfigDigest: "sha256:api-config", ContentDigest: "sha256:api-content"},
	}

	digest := semanticReleaseDigest("config-1", "sha256:runtime-v1", members)
	require.Equal(t, digest, semanticReleaseDigest("config-1", "sha256:runtime-v1", []app.AppReleaseMember{members[1], members[0]}))
	require.NotEqual(t, digest, semanticReleaseDigest("config-1", "sha256:runtime-v2", members))
}

func TestCanonicalDefinitionTOML(t *testing.T) {
	definition := map[string]any{
		"dependencies": []string{"database"},
		"helm": map[string]any{
			"chart_name": "api",
			"values":     map[string]any{"replicas": int64(2)},
		},
	}

	encoded, err := canonicalDefinitionTOML(definition)
	require.NoError(t, err)
	require.Equal(t, "dependencies = ['database']\n\n[helm]\nchart_name = 'api'\n\n[helm.values]\nreplicas = 2\n", encoded)
}

func TestReleaseDefinitionsBlobPrefersAuthoredTOML(t *testing.T) {
	sourceArchive := &pkgconfig.SourceArchive{Files: map[string]string{"policies/pass.rego": "package policies\n"}}
	blob, err := newReleaseDefinitionsBlob([]app.AppReleaseMember{
		{Kind: "component", LogicalName: "api", ConfigTOML: "type = 'helm_chart'\n"},
		{Kind: "sandbox", LogicalName: "sandbox", ConfigTOML: "type = 'aws-eks'\n"},
	}, map[string]string{"component:api": "# authored\nname = \"api\"\ntype = \"helm_chart\"\n"}, sourceArchive)
	require.NoError(t, err)

	var definitions releasearchive.ReleaseArchive
	require.NoError(t, json.Unmarshal([]byte(blob.String()), &definitions))
	require.Equal(t, 2, definitions.SchemaVersion)
	require.Equal(t, "# authored\nname = \"api\"\ntype = \"helm_chart\"\n", definitions.Members["component:api"])
	require.Equal(t, "type = 'aws-eks'\n", definitions.Members["sandbox:sandbox"])
	require.Equal(t, "package policies\n", definitions.Files["policies/pass.rego"])
}

func TestAppendAuthoredReleaseMembersAddsConfigurationWithoutDuplicatingLogicalMembers(t *testing.T) {
	archive := &pkgconfig.SourceArchive{
		Files: map[string]string{
			"components/api.toml":        "name = \"api\"\ntype = \"helm_chart\"\n",
			"metadata.toml":              "version = \"v2\"\n",
			"permissions/provision.toml": "type = \"provision\"\n",
			"policies/pass.rego":         "package policies\n",
		},
		Members: map[string]string{
			"component:api":        "components/api.toml",
			"metadata:metadata":    "metadata.toml",
			"permission:provision": "permissions/provision.toml",
		},
	}
	members, err := appendAuthoredReleaseMembers([]app.AppReleaseMember{{Kind: "component", LogicalName: "api", ContentDigest: "sha256:build"}}, archive)
	require.NoError(t, err)
	require.Len(t, members, 4)
	require.Equal(t, []string{"component:api", "metadata:metadata", "permission:provision", "source_file:policies/pass.rego"}, []string{
		releaseMemberKey(members[0]), releaseMemberKey(members[1]), releaseMemberKey(members[2]), releaseMemberKey(members[3]),
	})
	require.Equal(t, "version = \"v2\"\n", members[1].ConfigTOML)
	require.Equal(t, map[string]any{"path": "metadata.toml"}, members[1].SourceIdentity)
}
