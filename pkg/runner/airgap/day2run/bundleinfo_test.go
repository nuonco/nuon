package day2run

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
)

func TestBundleInfoFromManifest(t *testing.T) {
	manifest := bundle.LogicalManifest{
		Target: bundle.Target{OS: "linux", Architecture: "arm64"},
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

	require.Equal(t, day2.SchemaVersion, info.SchemaVersion)
	require.Equal(t, "dep", info.DeploymentID)
	require.Equal(t, "sha256:bundle", info.BundleDigest)
	require.Equal(t, activated, info.ActivatedAt)
	require.Equal(t, &day2.BundleTarget{OS: "linux", Architecture: "arm64"}, info.Target)
	require.True(t, info.Verification.BlobsVerified)
	require.True(t, info.Verification.EnvelopeParsed)
	require.Equal(t, int64(100+50+200+10+5+1000+2000), info.TotalSize)

	kinds := map[string]int{}
	for _, item := range info.Contents {
		kinds[item.Kind]++
	}
	require.Equal(t, map[string]int{
		day2.BundleContentKindComponent:    1,
		day2.BundleContentKindSandbox:      1,
		day2.BundleContentKindImage:        1,
		day2.BundleContentKindAction:       1,
		day2.BundleContentKindRunbook:      1,
		day2.BundleContentKindStackAsset:   1,
		day2.BundleContentKindRunnerBinary: 1,
		day2.BundleContentKindRunnerImage:  1,
	}, kinds)
	require.Equal(t, map[string]any{"helm": map[string]any{"chart_name": "api"}}, info.Contents[0].ComponentDefinition)
	require.Equal(t, &day2.BundleActionDefinition{TimeoutNanos: 30, Role: "operator", EnableKubeConfig: true,
		Triggers: []day2.BundleActionTrigger{{Type: "cron", CronSchedule: "0 * * * *"}}, Steps: []day2.BundleActionStep{{
			Name: "rollout", Index: 1, Command: "kubectl rollout restart", InlineContentsDigest: "sha256:inline", ArtifactDigest: "sha256:action", Environment: map[string]string{"TOKEN": "sha256:value"},
			Source: &day2.BundleSource{Repository: "nuonco/actions", RequestedRef: "main", Commit: "abc123", Directory: "restart", Version: "v1", Digest: "sha256:source"},
		}}}, info.Contents[3].ActionDefinition)
	require.Equal(t, &day2.BundleRunbookDefinition{Steps: []day2.BundleRunbookStep{{Kind: "action", Reference: "action:restart"}}}, info.Contents[4].RunbookDefinition)
}

func TestBundleInfoPublication(t *testing.T) {
	d, m, s, ex := testDispatcher(t)
	first := day2.BundleInfo{
		SchemaVersion: day2.SchemaVersion, DeploymentID: "dep", BundleDigest: "sha256:one",
		ActivatedAt: time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC),
	}
	local := map[string][]byte{}
	c, err := NewController(ControllerConfig{
		Mailbox: m, Envelope: d.envelope, Digest: "digest", DeploymentID: "dep", Owner: "owner",
		Executor: ex, Bundle: &first,
		WriteLocal: func(k string, b []byte) error { local[k] = b; return nil },
	})
	require.NoError(t, err)
	require.NoError(t, c.publishBundleInfo(context.Background()))

	var active day2.BundleInfo
	require.NoError(t, json.Unmarshal(local[day2.BundleKey], &active))
	require.Equal(t, "sha256:one", active.BundleDigest)
	require.NotEmpty(t, local[day2.BundleHistoryKey("sha256:one")])
	_, ok := s.objects["p/"+day2.BundleKey]
	require.True(t, ok)
	historyRaw, ok := s.objects["p/"+day2.BundleHistoryKey("sha256:one")]
	require.True(t, ok)

	// A restart republishes the active pointer but must not rewrite history,
	// and the active pointer adopts the first activation time from history
	// instead of the restart time.
	restarted := first
	restarted.ActivatedAt = first.ActivatedAt.Add(time.Hour)
	c.cfg.Bundle = &restarted
	require.NoError(t, c.publishBundleInfo(context.Background()))
	var kept day2.BundleInfo
	require.NoError(t, json.Unmarshal(s.objects["p/"+day2.BundleHistoryKey("sha256:one")].body, &kept))
	require.Equal(t, first.ActivatedAt, kept.ActivatedAt)
	require.Equal(t, historyRaw.etag, s.objects["p/"+day2.BundleHistoryKey("sha256:one")].etag)

	var republished day2.BundleInfo
	require.NoError(t, json.Unmarshal(s.objects["p/"+day2.BundleKey].body, &republished))
	require.Equal(t, first.ActivatedAt, republished.ActivatedAt)
	require.NoError(t, json.Unmarshal(local[day2.BundleKey], &republished))
	require.Equal(t, first.ActivatedAt, republished.ActivatedAt)
}

func TestBundleInfoPublicationSkippedWhenNil(t *testing.T) {
	d, m, s, ex := testDispatcher(t)
	c, err := NewController(ControllerConfig{Mailbox: m, Envelope: d.envelope, Digest: "digest", DeploymentID: "dep", Owner: "owner", Executor: ex})
	require.NoError(t, err)
	require.NoError(t, c.publishBundleInfo(context.Background()))
	_, ok := s.objects["p/"+day2.BundleKey]
	require.False(t, ok)
}
