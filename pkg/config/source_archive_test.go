package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	temporaldataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
)

func TestAppConfigSourceArchiveIsTemporalOnly(t *testing.T) {
	cfg := &AppConfig{
		Version: "v2",
		SourceArchive: &SourceArchive{
			SchemaVersion: 2,
			Files:         map[string]string{"metadata.toml": "authored source"},
		},
	}

	payload, err := temporaldataconverter.NewJSONConverter().ToPayload(cfg)
	require.NoError(t, err)
	var decoded AppConfig
	require.NoError(t, temporaldataconverter.NewJSONConverter().FromPayload(payload, &decoded))
	require.Equal(t, cfg.SourceArchive, decoded.SourceArchive)

	intermediate, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NotContains(t, string(intermediate), "authored source")
}

func TestSourceArchiveReindexMembersIgnoresNonTOMLFilesInDefinitionDirectories(t *testing.T) {
	archive := SourceArchive{Files: map[string]string{
		"permissions/provision.toml": "type = \"provision\"\n",
		"permissions/boundary.json":  "{}",
	}}

	require.NoError(t, archive.ReindexMembers())
	require.Equal(t, map[string]string{
		"permission:provision": "permissions/provision.toml",
	}, archive.Members)
}

func TestSourceArchiveReindexMembersUpgradesStoredIndexes(t *testing.T) {
	archive := SourceArchive{
		Files: map[string]string{
			"metadata.toml":              "version = \"v2\"\n",
			"permissions/provision.toml": "type = \"provision\"\n",
		},
		Members: map[string]string{},
	}

	require.NoError(t, archive.ReindexMembers())
	require.Equal(t, map[string]string{
		"metadata:metadata":    "metadata.toml",
		"permission:provision": "permissions/provision.toml",
	}, archive.Members)
}

func TestSourceArchiveV3TrustsStoredMemberIndex(t *testing.T) {
	archive := SourceArchive{
		SchemaVersion: 3,
		Files: map[string]string{
			"actions/renamed.toml": "name = \"different-from-index\"\n",
		},
		Members: map[string]string{
			"action:stable-name": "actions/renamed.toml",
		},
	}

	require.NoError(t, archive.ReindexMembers())
	require.Equal(t, "actions/renamed.toml", archive.Members["action:stable-name"])

	archive.Members["action:missing"] = "actions/missing.toml"
	require.ErrorContains(t, archive.ReindexMembers(), "references missing file")
}
