package prune

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)

func old() time.Time { return now.Add(-400 * 24 * time.Hour) }

func TestDecide(t *testing.T) {
	tests := map[string]struct {
		input         Input
		wantBatches   [][]string
		wantKept      map[string]KeepReason
		wantTruncated bool
	}{
		"old owned index is deleted before its children": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "idx", Tags: []string{"img-api-1.0"}, PushedAt: old(), Children: []string{"amd", "arm"}},
					{Digest: "amd", Tags: []string{"img-api-1.0--linux-amd64"}, PushedAt: old()},
					{Digest: "arm", PushedAt: old()},
				},
				OwnedTags: []string{"img-api-1.0"},
			},
			wantBatches: [][]string{{"idx"}, {"amd", "arm"}},
			wantKept:    map[string]KeepReason{},
		},
		"unowned tags are never touched": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "a", Tags: []string{"customer-pushed"}, PushedAt: old()},
				},
			},
			wantKept: map[string]KeepReason{"a": KeepReasonUnowned},
		},
		"keep tag keeps index and children": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "idx", Tags: []string{"img-api-1.0"}, PushedAt: old(), Children: []string{"amd"}},
					{Digest: "amd", Tags: []string{"img-api-1.0--linux-amd64"}, PushedAt: old()},
				},
				OwnedTags: []string{"img-api-1.0"},
				KeepTags:  []string{"img-api-1.0"},
			},
			wantKept: map[string]KeepReason{"idx": KeepReasonKeepTag, "amd": KeepReasonKeepTag},
		},
		"pinned old version kept by keep tag while newer unused version goes": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "pinned", Tags: []string{"img-api-0.19.1271"}, PushedAt: old()},
					{Digest: "unused", Tags: []string{"img-api-0.19.1250"}, PushedAt: old()},
				},
				OwnedTags: []string{"img-api-0.19.1271", "img-api-0.19.1250"},
				KeepTags:  []string{"img-api-0.19.1271"},
			},
			wantBatches: [][]string{{"unused"}},
			wantKept:    map[string]KeepReason{"pinned": KeepReasonKeepTag},
		},
		"images running in the cluster are kept by tag or digest": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "temporal", Tags: []string{"img-temporal-1.27.4"}, PushedAt: old()},
					{Digest: "clickhouse", Tags: []string{"img-clickhouse-24.3"}, PushedAt: old()},
				},
				OwnedTags:    []string{"img-temporal-1.27.4", "img-clickhouse-24.3"},
				InUseTags:    []string{"img-temporal-1.27.4"},
				InUseDigests: []string{"clickhouse"},
			},
			wantKept: map[string]KeepReason{"temporal": KeepReasonInUse, "clickhouse": KeepReasonInUse},
		},
		"recent pushes are kept by grace period": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "new", Tags: []string{"bld1"}, PushedAt: now.Add(-time.Hour)},
				},
				OwnedTags:   []string{"bld1"},
				GracePeriod: 30 * 24 * time.Hour,
			},
			wantKept: map[string]KeepReason{"new": KeepReasonGrace},
		},
		"child shared with a kept index survives": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "keep-idx", Tags: []string{"img-api-2.0"}, PushedAt: old(), Children: []string{"shared"}},
					{Digest: "drop-idx", Tags: []string{"img-api-1.0"}, PushedAt: old(), Children: []string{"shared"}},
					{Digest: "shared", PushedAt: old()},
				},
				OwnedTags: []string{"img-api-1.0", "img-api-2.0"},
				KeepTags:  []string{"img-api-2.0"},
			},
			wantBatches: [][]string{{"drop-idx"}},
			wantKept:    map[string]KeepReason{"keep-idx": KeepReasonKeepTag, "shared": KeepReasonParent},
		},
		"referrer follows its subject": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "kept", Tags: []string{"bld1"}, PushedAt: old()},
					{Digest: "kept-sig", PushedAt: old(), Subject: "kept"},
					{Digest: "gone", Tags: []string{"bld2"}, PushedAt: old()},
					{Digest: "gone-sig", PushedAt: old(), Subject: "gone"},
				},
				OwnedTags: []string{"bld1", "bld2"},
				KeepTags:  []string{"bld1"},
			},
			wantBatches: [][]string{{"gone-sig"}, {"gone"}},
			wantKept:    map[string]KeepReason{"kept": KeepReasonKeepTag, "kept-sig": KeepReasonSubject},
		},
		"legacy deploy tags are owned by pattern": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "legacy", Tags: []string{"dpl0123456789abcdefghijklm"}, PushedAt: old()},
					{Digest: "lookalike", Tags: []string{"dpl-not-an-id"}, PushedAt: old()},
				},
				OwnedPattern: []string{`^dpl[0-9a-z]{23}$`},
			},
			wantBatches: [][]string{{"legacy"}},
			wantKept:    map[string]KeepReason{"lookalike": KeepReasonUnowned},
		},
		"index with a missing child is kept": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "idx", Tags: []string{"bld1"}, PushedAt: old(), Children: []string{"missing"}},
				},
				OwnedTags: []string{"bld1"},
			},
			wantKept: map[string]KeepReason{"idx": KeepReasonDanglers},
		},
		"delete cap truncates": {
			input: Input{
				Manifests: []Manifest{
					{Digest: "a", Tags: []string{"bld1"}, PushedAt: old()},
					{Digest: "b", Tags: []string{"bld2"}, PushedAt: old()},
					{Digest: "c", Tags: []string{"bld3"}, PushedAt: old()},
				},
				OwnedTags:  []string{"bld1", "bld2", "bld3"},
				MaxDeletes: 2,
			},
			wantBatches:   [][]string{{"a", "b"}},
			wantKept:      map[string]KeepReason{},
			wantTruncated: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			in := tc.input
			in.Now = now
			if in.MaxDeletes == 0 {
				in.MaxDeletes = 100
			}

			got, err := Decide(in)
			require.NoError(t, err)

			var batches [][]string
			for _, b := range got.Batches {
				var digests []string
				for _, d := range b {
					digests = append(digests, d.Digest)
				}
				batches = append(batches, digests)
			}
			assert.Equal(t, tc.wantBatches, batches)
			assert.Equal(t, tc.wantKept, got.Kept)
			assert.Equal(t, tc.wantTruncated, got.Truncated)
		})
	}
}

func TestDecideRejectsBadPattern(t *testing.T) {
	_, err := Decide(Input{OwnedPattern: []string{"("}, MaxDeletes: 1})
	require.Error(t, err)
}

func TestParseImageRef(t *testing.T) {
	const repo = "123.dkr.ecr.us-west-2.amazonaws.com/inst1"
	tests := map[string]struct {
		ref        string
		wantTag    string
		wantDigest string
		wantOK     bool
	}{
		"tag":             {ref: repo + ":img-api-1.0", wantTag: "img-api-1.0", wantOK: true},
		"digest":          {ref: repo + "@sha256:abc", wantDigest: "sha256:abc", wantOK: true},
		"tag and digest":  {ref: repo + ":v1@sha256:abc", wantTag: "v1", wantDigest: "sha256:abc", wantOK: true},
		"pullable prefix": {ref: "docker-pullable://" + repo + "@sha256:abc", wantDigest: "sha256:abc", wantOK: true},
		"other repo":      {ref: "123.dkr.ecr.us-west-2.amazonaws.com/inst10:v1"},
		"other registry":  {ref: "docker.io/library/nginx:1.0"},
		"bare repo":       {ref: repo},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tag, digest, ok := parseImageRef(repo, tc.ref)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantTag, tag)
			assert.Equal(t, tc.wantDigest, digest)
		})
	}
}
