package operationrun

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
)

func TestBundleInfoFromManifest(t *testing.T) {
	manifest := bundle.LogicalManifest{
		Release: bundle.ReleaseIdentity{ID: "release-1", Digest: "sha256:release"},
		Package: bundle.PackageIdentity{ID: "package-1", Digest: "sha256:package", Format: "portable-oci", Target: "linux/arm64"},
		Target:  bundle.Target{OS: "linux", Architecture: "arm64"},
		Components: []bundle.Component{
			{Name: "api", Type: "helm_chart", Definition: bundle.ComponentDefinition{"helm": map[string]any{"chart_name": "api"}}, Artifact: bundle.Artifact{Digest: "sha256:a", Size: 100}},
		},
		Sandbox: &bundle.Sandbox{Type: "aws-eks", Source: bundle.Source{Repository: "nuonco/sandboxes"}, Artifact: bundle.Artifact{Digest: "sha256:s", Size: 50}},
		Images:  []bundle.Image{{Name: "nginx", Repository: "docker.io/nginx", Artifact: bundle.Artifact{Digest: "sha256:i", Size: 200}}},
		Actions: []bundle.Action{{Name: "restart", ConfigDigest: "sha256:c", Definition: &bundle.ActionDefinition{
			TimeoutNanos: 30, Role: "operator", EnableKubeConfig: true,
			Triggers: []bundle.ActionTriggerDefinition{{Type: "cron", CronSchedule: "0 * * * *"}},
			Steps:    []bundle.ActionStepDefinition{{Name: "rollout", Index: 1, Command: "kubectl rollout restart", InlineContentsDigest: "sha256:inline", Environment: map[string]string{"TOKEN": "sha256:value"}}},
		}, Steps: []bundle.Step{{
			Name: "rollout", Command: "kubectl rollout restart", InlineContentsDigest: "sha256:inline",
			Source:   &bundle.Source{Repository: "nuonco/actions", RequestedRef: "main", Commit: "abc123", Directory: "restart", Version: "v1", Digest: "sha256:source"},
			Artifact: &bundle.Artifact{Digest: "sha256:action", Size: 10},
		}}}},
		Runbooks: []bundle.Runbook{{Name: "recover", ConfigDigest: "sha256:runbook", Definition: bundle.RunbookDefinition{Steps: []bundle.RunbookStepDefinition{{Kind: "action", Reference: "action:restart"}}}}},
		StackAssets: []bundle.StackAsset{
			{Role: "cloudformation", SourceURL: "https://example.com/stack.yaml", Digest: "sha256:t", Size: 5},
		},
		Runner: &bundle.Runner{
			Version: "v1.2.3",
			Binary:  &bundle.Artifact{Digest: "sha256:rb", Size: 1000},
			Image:   &bundle.Image{Name: "runner", Repository: "public.ecr.aws/nuon/runner", Artifact: bundle.Artifact{Digest: "sha256:ri", Size: 2000}},
		},
	}
	activated := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	info := BundleInfoFromManifest("dep", "sha256:bundle", manifest, activated)

	require.Equal(t, operation.SchemaVersion, info.SchemaVersion)
	require.Equal(t, "dep", info.DeploymentID)
	require.Equal(t, "sha256:bundle", info.BundleDigest)
	require.Equal(t, &operation.BundleReleaseIdentity{ID: "release-1", Digest: "sha256:release"}, info.Release)
	require.Equal(t, &operation.BundlePackageIdentity{ID: "package-1", Digest: "sha256:package", Format: "portable-oci", Target: "linux/arm64"}, info.Package)
	require.Equal(t, activated, info.ActivatedAt)
	require.Equal(t, &operation.BundleTarget{OS: "linux", Architecture: "arm64"}, info.Target)
	require.True(t, info.Verification.BlobsVerified)
	require.True(t, info.Verification.EnvelopeParsed)
	require.Equal(t, int64(100+50+200+10+5+1000+2000), info.TotalSize)

	kinds := map[string]int{}
	for _, item := range info.Contents {
		kinds[item.Kind]++
	}
	require.Equal(t, map[string]int{
		operation.BundleContentKindComponent:    1,
		operation.BundleContentKindSandbox:      1,
		operation.BundleContentKindImage:        1,
		operation.BundleContentKindAction:       1,
		operation.BundleContentKindRunbook:      1,
		operation.BundleContentKindStackAsset:   1,
		operation.BundleContentKindRunnerBinary: 1,
		operation.BundleContentKindRunnerImage:  1,
	}, kinds)
	require.Equal(t, map[string]any{"helm": map[string]any{"chart_name": "api"}}, info.Contents[0].ComponentDefinition)
	require.Equal(t, &operation.BundleActionDefinition{TimeoutNanos: 30, Role: "operator", EnableKubeConfig: true,
		Triggers: []operation.BundleActionTrigger{{Type: "cron", CronSchedule: "0 * * * *"}}, Steps: []operation.BundleActionStep{{
			Name: "rollout", Index: 1, Command: "kubectl rollout restart", InlineContentsDigest: "sha256:inline", ArtifactDigest: "sha256:action", Environment: map[string]string{"TOKEN": "sha256:value"},
			Source: &operation.BundleSource{Repository: "nuonco/actions", RequestedRef: "main", Commit: "abc123", Directory: "restart", Version: "v1", Digest: "sha256:source"},
		}}}, info.Contents[3].ActionDefinition)
	require.Equal(t, &operation.BundleRunbookDefinition{Steps: []operation.BundleRunbookStep{{Kind: "action", Reference: "action:restart"}}}, info.Contents[4].RunbookDefinition)
}

func TestBundleInfoPublication(t *testing.T) {
	d, m, s, ex := testDispatcher(t)
	first := operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "dep", BundleDigest: "sha256:one",
		OperationID: "install-run-1", ActivatedAt: time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC),
	}
	local := map[string][]byte{}
	c, err := NewController(ControllerConfig{
		Mailbox: m, Envelope: d.envelope, Digest: "digest", DeploymentID: "dep", Owner: "owner",
		Executor: ex, Bundle: &first,
		WriteLocal: func(k string, b []byte) error { local[k] = b; return nil },
	})
	require.NoError(t, err)
	require.NoError(t, c.publishBundleInfo(context.Background()))

	var active operation.BundleInfo
	require.NoError(t, json.Unmarshal(local[operation.BundleKey], &active))
	require.Equal(t, "sha256:one", active.BundleDigest)
	require.NotEmpty(t, local[operation.BundleHistoryKey("sha256:one")])
	_, ok := s.objects["p/"+operation.BundleKey]
	require.True(t, ok)
	historyRaw, ok := s.objects["p/"+operation.BundleHistoryKey("sha256:one")]
	require.True(t, ok)

	// A restart republishes the active pointer but must not rewrite history,
	// and the active pointer adopts the first activation time from history
	// instead of the restart time.
	restarted := first
	restarted.ActivatedAt = first.ActivatedAt.Add(time.Hour)
	restarted.OperationID = "restart-run"
	c.cfg.Bundle = &restarted
	require.NoError(t, c.publishBundleInfo(context.Background()))
	var kept operation.BundleInfo
	require.NoError(t, json.Unmarshal(s.objects["p/"+operation.BundleHistoryKey("sha256:one")].body, &kept))
	require.Equal(t, first.ActivatedAt, kept.ActivatedAt)
	require.Equal(t, historyRaw.etag, s.objects["p/"+operation.BundleHistoryKey("sha256:one")].etag)

	var republished operation.BundleInfo
	require.NoError(t, json.Unmarshal(s.objects["p/"+operation.BundleKey].body, &republished))
	require.Equal(t, first.ActivatedAt, republished.ActivatedAt)
	require.Equal(t, first.OperationID, republished.OperationID)
	require.NoError(t, json.Unmarshal(local[operation.BundleKey], &republished))
	require.Equal(t, first.ActivatedAt, republished.ActivatedAt)
	require.Equal(t, first.OperationID, republished.OperationID)
}

func TestBundleInfoPublicationSkippedWhenNil(t *testing.T) {
	d, m, s, ex := testDispatcher(t)
	c, err := NewController(ControllerConfig{Mailbox: m, Envelope: d.envelope, Digest: "digest", DeploymentID: "dep", Owner: "owner", Executor: ex})
	require.NoError(t, err)
	require.NoError(t, c.publishBundleInfo(context.Background()))
	_, ok := s.objects["p/"+operation.BundleKey]
	require.False(t, ok)
}
