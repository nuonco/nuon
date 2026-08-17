package cmd

import (
	"context"
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
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "job-logs"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "steps", "one", "result.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "job-logs", "one.ndjson"), []byte("{\"msg\":\"started\"}\n"), 0o600))

	files, err := collectUploadFiles(dir, "install/state", legacyRunnerStatePath)
	require.NoError(t, err)
	keys := make([]string, 0, len(files))
	for _, file := range files {
		keys = append(keys, file.key)
	}
	require.ElementsMatch(t, []string{
		"install/state/status.json",
		"install/state/steps/one/result.json",
		"install/state/job-logs/one.ndjson",
	}, keys)
}

func TestCollectUploadFilesSkipsReservedObjects(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "tfstate"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "runner"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "day2"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "LEASE"), []byte(`{"released":true}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DONE"), []byte("success"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "tfstate", "workspace.json.lock"), []byte(`{"ID":"stale"}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "runner", "heartbeat.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "day2", "candidate.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "day2", "staged-candidate.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "day2", "stack-candidate.json"), []byte("{}"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "future-coordination.json"), []byte("{}"), 0o600))

	files, err := collectUploadFiles(dir, "install/state", legacyRunnerStatePath)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, "install/state/status.json", files[0].key, "only runner-owned state paths may be uploaded")

	require.NoError(t, removeTerraformLocks(dir))
	_, err = os.Stat(filepath.Join(dir, "tfstate", "workspace.json.lock"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestLegacyRunnerStatePath(t *testing.T) {
	for _, name := range []string{
		"status.json",
		"report.json",
		"install-runs/run-1/events/0001.json",
		"install-controls/run-1/retry.handled.json",
		"steps/plan/result.json",
		"step-plans/plan.json",
		"job-logs/plan.ndjson",
		"health/latest.json",
		"tfstate/workspace.json",
		"runs/run-1/status.json",
		"job-plans/job-1.json",
		"schedules/action-1/cursor.json",
		"day2/catalog.json",
		"day2/bundle.json",
		"day2/bundles/sha256-one.json",
	} {
		require.Truef(t, legacyRunnerStatePath(name), "expected legacy runner ownership of %s", name)
	}
	for _, name := range []string{
		"DONE",
		"LEASE",
		"runner/heartbeat.json",
		"install-controls/run-1/retry.json",
		"dispatch/requests/request.json",
		"day2/candidate.json",
		"day2/staged-candidate.json",
		"day2/stack-candidate.json",
		"day2/candidates/sha256-one/approval.json",
		"future-coordination.json",
	} {
		require.Falsef(t, legacyRunnerStatePath(name), "expected external ownership of %s", name)
	}
}

func TestCollectUploadFilesWithoutOwnershipFilter(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "result.json"), []byte("{}"), 0o600))

	files, err := collectUploadFiles(dir, "runs/run-1", nil)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, "runs/run-1/result.json", files[0].key)
}

func TestUploadDirSyncsUnknownRunnerArtifact(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "future"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "future", "artifact.json"), []byte(`{"value":true}`), 0o600))
	client := newFakeLeaseS3()
	syncer := &airgapS3Sync{client: client, bucket: "state", prefix: "deployment"}

	require.NoError(t, syncer.uploadDir(context.Background(), dir))

	client.mu.Lock()
	defer client.mu.Unlock()
	require.Contains(t, client.objects, "deployment/runner/future/artifact.json")
}

func TestMigrateLegacyLocalRunnerStateRecoversJobLogs(t *testing.T) {
	root := t.TempDir()
	runnerDir := filepath.Join(root, "runner")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "job-logs"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "dispatch", "requests"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "status.json"), []byte(`{"status":"finished"}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "job-logs", "plan.ndjson"), []byte("log\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "dispatch", "requests", "one.json"), []byte("control"), 0o600))

	require.NoError(t, migrateLegacyLocalRunnerState(root, runnerDir))

	require.FileExists(t, filepath.Join(runnerDir, "status.json"))
	require.FileExists(t, filepath.Join(runnerDir, "job-logs", "plan.ndjson"))
	require.NoFileExists(t, filepath.Join(runnerDir, "dispatch", "requests", "one.json"))
}

func TestWriteRunnerHeartbeatOverwritesOneObject(t *testing.T) {
	client := newFakeLeaseS3()
	syncer := &airgapS3Sync{client: client, bucket: "state", prefix: "deployment"}

	require.NoError(t, syncer.writeRunnerHeartbeat(context.Background(), []byte(`{"observed_at":"first"}`)))
	require.NoError(t, syncer.writeRunnerHeartbeat(context.Background(), []byte(`{"observed_at":"second"}`)))

	client.mu.Lock()
	defer client.mu.Unlock()
	require.Len(t, client.objects, 1)
	require.JSONEq(t, `{"observed_at":"second"}`, string(client.objects["deployment/runner/heartbeat.json"].body))
}
