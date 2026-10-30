package helpers

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/sourceidentity"
)

func cloneBuildSource(i *sourceidentity.Identity) *sourceidentity.Identity {
	if i == nil {
		return nil
	}
	out := *i
	out.Files = slices.Clone(i.Files)
	out.Reasons = slices.Clone(i.Reasons)
	return &out
}

func drawBuildSource(t *hegel.T) *sourceidentity.Identity {
	i := sourceidentity.New(hegel.Draw(t, hegel.SampledFrom([]string{"terraform", "helm", "image"})),
		sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(16))))
	i.Root = hegel.Draw(t, hegel.SampledFrom([]string{"", "src", "nested/é"}))
	i.Repository = "https://example.com/acme/repo"
	i.Commit = strings.Repeat("a", 40)
	i.Digest = "sha256:" + sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(16)))
	i.Files = []sourceidentity.File{
		{Path: "z.txt", SHA256: sourceidentity.Hash("z")},
		{Path: "a.txt", SHA256: sourceidentity.Hash("a")},
	}
	i.MarkUnknown("configured_file_inputs_uninspected")
	for _, reason := range hegel.Draw(t, hegel.Lists(hegel.SampledFrom([]string{"external_dependency", "runtime_inputs"})).MaxSize(4)) {
		i.MarkUnknown(reason)
	}
	i.Seal()
	return i
}

func TestBuildSourceClaimKeepsSelection(t *testing.T) {
	for _, field := range []string{"root", "digest"} {
		t.Run(field, func(t *testing.T) {
			hegel.Test(t, func(t *hegel.T) {
				expected := drawBuildSource(t)
				actual := cloneBuildSource(expected)
				actual.Captured = true
				if field == "root" {
					actual.Root += "/other"
				} else {
					actual.Digest = "sha256:" + sourceidentity.Hash(actual.Digest)
				}
				actual.Seal()
				require.NoError(t, actual.Validate())
				require.Error(t, sourceidentity.ValidateTransition(expected, actual, true), field)
			})
		})
	}
}

func TestReportedBuildSourceDoesNotSortPreparedAliases(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		expected := drawBuildSource(t)
		expected.Captured = hegel.Draw(t, hegel.Booleans())
		slices.Reverse(expected.Files)
		alias := *expected
		before := cloneBuildSource(expected)
		require.NoError(t, expected.Validate())
		actual := reportedBuildSource(expected, nil)
		require.Equal(t, before, expected)
		require.Equal(t, before.Files, alias.Files)
		require.Equal(t, before.Reasons, alias.Reasons)
		require.NoError(t, actual.Validate())
		require.Empty(t, actual.EligibleChecksum())
		require.Equal(t, before.Checksum, actual.Checksum)
		require.Contains(t, actual.Reasons, "runner_source_evidence_missing")
		for _, reason := range before.Reasons {
			require.Contains(t, actual.Reasons, reason)
		}
		actual.Files[0].Path = "mutated"
		actual.Reasons[0] = "mutated"
		require.Equal(t, before, expected)
	})
}

func TestBuildSourcePreparedClaimedReportedModel(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		base := drawBuildSource(t)
		for _, captured := range []bool{false, true} {
			prepared := cloneBuildSource(base)
			prepared.Captured = captured
			prepared.MarkUnknown("source_requires_runner_resolution")
			claimed := cloneBuildSource(prepared)
			claimed.Captured = true
			if !captured {
				claimed.Reasons = slices.DeleteFunc(claimed.Reasons, func(s string) bool { return s == "source_requires_runner_resolution" })
			}
			claimed.Seal()
			before := cloneBuildSource(prepared)
			require.NoError(t, sourceidentity.ValidateTransition(prepared, claimed, true))
			require.Equal(t, captured, sourceidentity.ValidateTransition(prepared, claimed, false) == nil)
			require.NoError(t, sourceidentity.ValidateTransition(claimed, claimed, false))
			for _, mutation := range []string{"config", "kind", "root", "digest", "repository", "commit", "files", "reasons", "captured", "integrity"} {
				changed := cloneBuildSource(claimed)
				switch mutation {
				case "config":
					changed.ConfigChecksum += "!"
				case "kind":
					changed.Kind += "!"
				case "root":
					changed.Root += "/other"
				case "digest":
					changed.Digest = "sha256:" + sourceidentity.Hash(changed.Digest)
				case "repository":
					changed.Repository += "/other"
				case "commit":
					changed.Commit = strings.Repeat("b", 40)
				case "files":
					changed.Files[0].Executable = !changed.Files[0].Executable
				case "reasons":
					changed.Reasons = []string{"replacement_unknown"}
				case "captured":
					changed.Captured = false
				}
				changed.Seal()
				if mutation == "integrity" {
					changed.Checksum += "!"
				}
				require.Error(t, sourceidentity.ValidateTransition(claimed, changed, true), mutation)
				require.Error(t, sourceidentity.ValidateTransition(claimed, changed, false), mutation)
				allowedInitial := !captured && mutation == "files"
				require.Equal(t, allowedInitial, sourceidentity.ValidateTransition(prepared, changed, true) == nil, mutation)
			}
			require.NoError(t, sourceidentity.ValidateTransition(prepared, nil, false))
			missing := reportedBuildSource(prepared, nil)
			require.Empty(t, missing.EligibleChecksum())
			require.Contains(t, missing.Reasons, "runner_source_evidence_missing")
			data, err := json.Marshal(claimed)
			require.NoError(t, err)
			var transported sourceidentity.Identity
			require.NoError(t, json.Unmarshal(data, &transported))
			require.NoError(t, sourceidentity.ValidateTransition(claimed, &transported, false))
			require.Equal(t, before, prepared)
		}
		require.NoError(t, sourceidentity.ValidateTransition(nil, nil, false))
		require.Nil(t, reportedBuildSource(nil, nil))
		require.Error(t, sourceidentity.ValidateTransition(nil, base, true))
	})
}

func TestBuildSourceUnpreparedClosureCannotBecomeComplete(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		for _, kind := range []string{"terraform", "helm", "kustomize", "pulumi", "image"} {
			expected := drawBuildSource(t)
			expected.Kind = kind
			expected.Reasons = []string{"source_requires_runner_resolution"}
			expected.Seal()
			before := cloneBuildSource(expected)
			actual := cloneBuildSource(expected)
			actual.Captured, actual.Coverage, actual.Reasons = true, sourceidentity.Complete, nil
			actual.Seal()
			require.NoError(t, actual.Validate())
			require.Equal(t, kind == "image", sourceidentity.ValidateTransition(expected, actual, true) == nil, kind)
			require.Error(t, sourceidentity.ValidateTransition(expected, actual, false), kind)
			require.Equal(t, before, expected)
		}
	})
}
