package customermanaged

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReleaseArchiveFileListIsSortedWithContentMetadata(t *testing.T) {
	archive := ReleaseArchive{Files: map[string]string{
		"policies/pass.rego": "package policies\n",
		"metadata.toml":      "name = \"example\"\n",
	}}

	files := archive.FileList()
	require.Len(t, files, 2)
	require.Equal(t, "metadata.toml", files[0].Path)
	require.Equal(t, "application/toml", files[0].MediaType)
	require.Equal(t, int64(len("name = \"example\"\n")), files[0].Size)
	require.Contains(t, files[0].Digest, "sha256:")
	require.Equal(t, "policies/pass.rego", files[1].Path)
	require.Equal(t, "text/x-rego", files[1].MediaType)
}
