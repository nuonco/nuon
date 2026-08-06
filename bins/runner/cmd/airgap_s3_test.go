package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseS3URI(t *testing.T) {
	tests := []struct {
		name, uri, bucket, key string
		wantErr                bool
	}{
		{name: "object", uri: "s3://bucket/path/to/object", bucket: "bucket", key: "path/to/object"},
		{name: "bucket", uri: "s3://bucket", bucket: "bucket"},
		{name: "trim slashes", uri: "s3://bucket/prefix/", bucket: "bucket", key: "prefix"},
		{name: "not s3", uri: "/tmp/file", wantErr: true},
		{name: "missing bucket", uri: "s3:///key", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := parseS3URI(tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.bucket, bucket)
			require.Equal(t, tt.key, key)
		})
	}
}

func TestCollectUploadFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "steps", "one"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "steps", "one", "result.json"), []byte("{}"), 0o600))

	files, err := collectUploadFiles(dir, "install/state")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"install/state/status.json", "install/state/steps/one/result.json"}, []string{files[0].key, files[1].key})
}

func TestCollectUploadFilesSkipsReservedObjects(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "tfstate"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "LEASE"), []byte(`{"released":true}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DONE"), []byte("success"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "tfstate", "workspace.json.lock"), []byte(`{"ID":"stale"}`), 0o600))

	files, err := collectUploadFiles(dir, "install/state")
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, "install/state/status.json", files[0].key, "coordination and ephemeral Terraform lock files must not be uploaded")

	require.NoError(t, removeTerraformLocks(dir))
	_, err = os.Stat(filepath.Join(dir, "tfstate", "workspace.json.lock"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
