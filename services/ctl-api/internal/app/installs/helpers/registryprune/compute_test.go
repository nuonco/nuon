package registryprune

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)

func daysAgo(n int) time.Time { return now.Add(-time.Duration(n) * 24 * time.Hour) }

func TestCompute(t *testing.T) {
	tests := map[string]struct {
		input     ComputeInput
		wantOwned []string
		wantKeep  []string
	}{
		"keeps the newest K distinct tags per component": {
			input: ComputeInput{
				Deploys: []Deploy{
					{ID: "d1", ComponentID: "api", Tag: "img-api-1", CreatedAt: daysAgo(90)},
					{ID: "d2", ComponentID: "api", Tag: "img-api-2", CreatedAt: daysAgo(80)},
					{ID: "d3", ComponentID: "api", Tag: "img-api-2", CreatedAt: daysAgo(70)},
					{ID: "d4", ComponentID: "api", Tag: "img-api-3", CreatedAt: daysAgo(60)},
				},
				DefaultCount: 2,
			},
			wantOwned: []string{"img-api-1", "img-api-2", "img-api-3"},
			wantKeep:  []string{"img-api-2", "img-api-3"},
		},
		"component override beats default": {
			input: ComputeInput{
				Deploys: []Deploy{
					{ID: "d1", ComponentID: "temporal", Tag: "img-temporal-1", CreatedAt: daysAgo(400)},
					{ID: "d2", ComponentID: "temporal", Tag: "img-temporal-2", CreatedAt: daysAgo(300)},
				},
				RetainCounts: map[string]int{"temporal": 10},
				DefaultCount: 1,
			},
			wantOwned: []string{"img-temporal-1", "img-temporal-2"},
			wantKeep:  []string{"img-temporal-1", "img-temporal-2"},
		},
		"pinned version referenced by a live helm plan is kept past K": {
			input: ComputeInput{
				Deploys: []Deploy{
					{ID: "d1", ComponentID: "img-api", Tag: "img-nuon-ctl-api-0.19.1271", CreatedAt: daysAgo(200)},
					{ID: "d2", ComponentID: "img-api", Tag: "img-nuon-ctl-api-0.19.1280", CreatedAt: daysAgo(100)},
					{ID: "d3", ComponentID: "img-api", Tag: "img-nuon-ctl-api-0.19.1298", CreatedAt: daysAgo(90)},
				},
				DefaultCount: 1,
				LivePlans:    []string{`{"values":{"image":{"tag":"img-nuon-ctl-api-0.19.1271"}}}`},
			},
			wantOwned: []string{"img-nuon-ctl-api-0.19.1271", "img-nuon-ctl-api-0.19.1280", "img-nuon-ctl-api-0.19.1298"},
			wantKeep:  []string{"img-nuon-ctl-api-0.19.1271", "img-nuon-ctl-api-0.19.1298"},
		},
		"live, in-flight and recent deploys are kept with their legacy deploy tags": {
			input: ComputeInput{
				Deploys: []Deploy{
					{ID: "live", ComponentID: "a", Tag: "bld-live", CreatedAt: daysAgo(500), Live: true},
					{ID: "pending", ComponentID: "a", Tag: "bld-pending", CreatedAt: daysAgo(400), InFlight: true},
					{ID: "recent", ComponentID: "a", Tag: "bld-recent", CreatedAt: daysAgo(3)},
					{ID: "stale", ComponentID: "a", Tag: "bld-stale", CreatedAt: daysAgo(300)},
				},
				DefaultCount: 1,
				GracePeriod:  30 * 24 * time.Hour,
			},
			wantOwned: []string{"bld-live", "bld-pending", "bld-recent", "bld-stale"},
			wantKeep:  []string{"bld-live", "bld-pending", "bld-recent", "live", "pending", "recent"},
		},
		"rolled forward but never deployed still counts toward K": {
			input: ComputeInput{
				Deploys: []Deploy{
					{ID: "d1", ComponentID: "api", Tag: "img-api-good", CreatedAt: daysAgo(60), Live: true},
					{ID: "d2", ComponentID: "api", Tag: "img-api-broken", CreatedAt: daysAgo(50)},
				},
				DefaultCount: 1,
			},
			wantOwned: []string{"img-api-broken", "img-api-good"},
			wantKeep:  []string{"d1", "img-api-broken", "img-api-good"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			in := tc.input
			in.Now = now
			got := Compute(in)
			assert.Equal(t, tc.wantOwned, got.Owned)
			assert.Equal(t, tc.wantKeep, got.Keep)
		})
	}
}

func TestInstallRegistryTag(t *testing.T) {
	tests := map[string]struct {
		component, build, digest, resolved string
		shared                             bool
		want                               string
	}{
		"artifact without source digest uses build id": {component: "api", build: "bld1", want: "bld1"},
		"image pinned by digest uses build id":         {component: "api", build: "bld1", digest: "sha256:x", want: "bld1"},
		"image in shared repo is prefixed":             {component: "img_nuon_ctl_api", build: "bld1", digest: "sha256:x", resolved: "0.19.1280", shared: true, want: "img-nuon-ctl-api-0.19.1280"},
		"image in per-component repo is bare":          {component: "api", build: "bld1", digest: "sha256:x", resolved: "1.0", want: "1.0"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, InstallRegistryTag(tc.component, tc.build, tc.digest, tc.resolved, tc.shared))
		})
	}
}
