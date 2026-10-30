package buildinputs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourcefs"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func cloneSource(i *sourceidentity.Identity) *sourceidentity.Identity {
	if i == nil {
		return nil
	}
	out := *i
	out.Files = slices.Clone(i.Files)
	out.Reasons = slices.Clone(i.Reasons)
	return &out
}

func drawImageSource(t *hegel.T) *sourceidentity.Identity {
	i := sourceidentity.New("image", "config:"+hegel.Draw(t, hegel.Text().MaxSize(16)))
	i.Root = hegel.Draw(t, hegel.SampledFrom([]string{"src", "nested/é"}))
	i.Repository = "https://example.com/acme/repo"
	i.Commit = strings.Repeat("a", 40)
	i.MarkUnknown("source_requires_runner_resolution")
	i.MarkUnknown("configured_file_inputs_uninspected")
	i.Seal()
	return i
}

func TestImageResolutionPreservesPreparedEvidence(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		expected := drawImageSource(t)
		before := cloneSource(expected)
		alias := *expected
		digest := "sha256:" + sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(20)))
		actual, err := sourceidentity.ResolveImage(expected, digest)
		require.NoError(t, err)
		want := cloneSource(before)
		want.Captured, want.Digest = true, digest
		want.Reasons = []string{"configured_file_inputs_uninspected"}
		want.Seal()
		require.Equal(t, want, actual)
		require.NoError(t, actual.Validate())
		require.Equal(t, before, expected)
		require.Equal(t, before.Reasons, alias.Reasons)
	})
}

func TestImageResolutionRejectsInvalidAndPinnedDigests(t *testing.T) {
	for _, mode := range []string{"invalid", "pinned", "kind"} {
		t.Run(mode, func(t *testing.T) {
			hegel.Test(t, func(t *hegel.T) {
				expected := drawImageSource(t)
				digest := "sha256:" + sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(20)))
				switch mode {
				case "invalid":
					digest = hegel.Draw(t, hegel.SampledFrom([]string{"", "latest", "sha256:abc", "sha256:" + strings.Repeat("g", 64)}))
				case "pinned":
					expected.Digest = "sha256:" + sourceidentity.Hash(digest)
				case "kind":
					expected.Kind = "terraform"
				}
				expected.Seal()
				before := cloneSource(expected)
				actual, err := sourceidentity.ResolveImage(expected, digest)
				require.Error(t, err)
				require.Nil(t, actual)
				require.Equal(t, before, expected)
			})
		})
	}
}

type receiptClient struct {
	call func(context.Context, string, string, *models.ServiceCreateRunnerJobExecutionResultRequest) error
}

func (c receiptClient) CreateJobExecutionResult(ctx context.Context, job, execution string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	return nil, c.call(ctx, job, execution, req)
}

func TestClaimAndReportTransportProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		i := drawImageSource(t)
		i.Captured = hegel.Draw(t, hegel.Booleans())
		i.Digest = "sha256:" + sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(16)))
		if hegel.Draw(t, hegel.Booleans()) {
			i.Coverage, i.Reasons = sourceidentity.Complete, nil
		}
		i.Files = []sourceidentity.File{{Path: "é.txt", SHA256: sourceidentity.Hash("contents"), Executable: true}}
		i.Seal()
		before := cloneSource(i)
		alias := *i
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		transportErr := errors.New("receipt transport failed")
		var wantErr error
		if hegel.Draw(t, hegel.Booleans()) {
			wantErr = transportErr
		}
		for _, claim := range []bool{true, false} {
			calls := 0
			client := receiptClient{call: func(actualCtx context.Context, job, execution string, req *models.ServiceCreateRunnerJobExecutionResultRequest) error {
				calls++
				require.Equal(t, ctx, actualCtx)
				require.Equal(t, "job", job)
				require.Equal(t, "execution", execution)
				metadata := req.SourceIdentity
				if claim {
					require.Nil(t, metadata)
					require.False(t, req.Success)
					metadata = req.SourceCapture
				} else {
					require.Nil(t, req.SourceCapture)
					require.True(t, req.Success)
					require.Equal(t, int64(17), req.ErrorCode)
				}
				data, err := json.Marshal(metadata)
				require.NoError(t, err)
				var received sourceidentity.Identity
				require.NoError(t, json.Unmarshal(data, &received))
				require.Equal(t, before, &received)
				metadata.Files[0].Path = "mutated"
				if len(metadata.Reasons) > 0 {
					metadata.Reasons[0] = "mutated"
				}
				return wantErr
			}}
			var err error
			if claim {
				err = Claim(ctx, client, "job", "execution", i)
			} else {
				err = Report(ctx, client, "job", "execution", &models.ServiceCreateRunnerJobExecutionResultRequest{Success: true, ErrorCode: 17}, i)
			}
			require.ErrorIs(t, err, wantErr)
			require.Equal(t, 1, calls)
			require.Equal(t, before, i)
			require.Equal(t, before.Files, alias.Files)
			require.Equal(t, before.Reasons, alias.Reasons)
		}
		calls := 0
		client := receiptClient{call: func(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) error {
			calls++
			require.Nil(t, req.SourceIdentity)
			require.Nil(t, req.SourceCapture)
			return nil
		}}
		require.NoError(t, Claim(ctx, client, "job", "execution", nil))
		require.Zero(t, calls)
		require.NoError(t, Report(ctx, client, "job", "execution", &models.ServiceCreateRunnerJobExecutionResultRequest{}, nil))
		require.Equal(t, 1, calls)
		invalid := cloneSource(i)
		invalid.Checksum += "!"
		require.Error(t, Claim(ctx, client, "job", "execution", invalid))
		require.Error(t, Report(ctx, client, "job", "execution", &models.ServiceCreateRunnerJobExecutionResultRequest{}, invalid))
		require.Equal(t, 1, calls)
	})
}

func TestVerifyPreparedGitSelectionAndDriftProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fs, root := afero.NewMemMapFs(), "/repo"
		require.NoError(t, fs.MkdirAll(filepath.Join(root, "src"), 0755))
		data := hegel.Draw(t, hegel.Text().MaxSize(24))
		file := filepath.Join(root, "src", "input.txt")
		require.NoError(t, afero.WriteFile(fs, file, []byte(data), 0644))
		i := sourceidentity.New("terraform", "config")
		i.Root, i.Commit = "src", strings.Repeat("a", 40)
		if hegel.Draw(t, hegel.Booleans()) {
			i.MarkUnknown("external_dependency")
		}
		source, err := sourcefs.New(fs, root)
		require.NoError(t, err)
		i, err = sourceidentity.Capture(source, i)
		require.NoError(t, err)
		before := cloneSource(i)
		plan := &plantypes.BuildPlan{SourceIdentity: i, Src: &plantypes.GitSource{Ref: i.Commit, Path: "src/."}}
		require.NoError(t, Verify(plan, source))
		for _, ref := range []string{"main", strings.Repeat("b", 40)} {
			plan.Src.Ref = ref
			require.Error(t, Verify(plan, source))
		}
		plan.Src.Ref, plan.Src.Path = i.Commit, "other"
		require.Error(t, Verify(plan, source))
		plan.Src.Path = "src"
		require.NoError(t, afero.WriteFile(fs, file, []byte(data+"!"), 0644))
		require.Error(t, Verify(plan, source))
		require.Equal(t, before, i)
		uncaptured := cloneSource(i)
		uncaptured.Captured = false
		uncaptured.MarkUnknown("source_requires_runner_resolution")
		plan.SourceIdentity = uncaptured
		require.NoError(t, Verify(plan, nil))
		require.Empty(t, uncaptured.EligibleChecksum())
		plan.Src.Ref = "main"
		require.Error(t, Verify(plan, nil))
		plan.SourceIdentity = nil
		require.NoError(t, Verify(plan, nil))
	})
}

func TestVerifyInlineDigestProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		manifest := hegel.Draw(t, hegel.Text().MaxSize(40))
		data, err := json.Marshal(manifest)
		require.NoError(t, err)
		digest := sha256.Sum256(data)
		i := sourceidentity.New("inline", "config")
		i.Captured, i.Digest = true, hex.EncodeToString(digest[:])
		i.Seal()
		before := cloneSource(i)
		plan := &plantypes.BuildPlan{SourceIdentity: i, KubernetesManifestBuildPlan: &plantypes.KubernetesManifestBuildPlan{SourceType: "inline", InlineManifest: manifest}}
		require.NoError(t, Verify(plan, nil))
		plan.KubernetesManifestBuildPlan.InlineManifest += "!"
		require.Error(t, Verify(plan, nil))
		require.Equal(t, before, i)
	})
}

func TestImageResolutionCompletenessAndRetryProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		expected := drawImageSource(t)
		expected.Reasons = []string{"source_requires_runner_resolution"}
		expected.Seal()
		digest := "sha256:" + sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(16)))
		resolved, err := sourceidentity.ResolveImage(expected, digest)
		require.NoError(t, err)
		require.Equal(t, sourceidentity.Complete, resolved.Coverage)
		require.Empty(t, resolved.Reasons)
		require.Equal(t, resolved.Checksum, resolved.EligibleChecksum())
		retried, err := sourceidentity.ResolveImage(resolved, digest)
		require.NoError(t, err)
		require.Equal(t, resolved, retried)
		retried, err = sourceidentity.ResolveImage(resolved, "sha256:"+sourceidentity.Hash(digest))
		require.Error(t, err)
		require.Nil(t, retried)
		missing, err := sourceidentity.ResolveImage(nil, digest)
		require.NoError(t, err)
		require.Nil(t, missing)
	})
}
