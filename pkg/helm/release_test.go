package helm

import (
	"testing"

	release "helm.sh/helm/v4/pkg/release/v1"
)

func rev(version int, status release.Status) *release.Release {
	return &release.Release{
		Version: version,
		Info:    &release.Info{Status: status},
	}
}

func TestIsPending(t *testing.T) {
	tests := []struct {
		name string
		rel  *release.Release
		want bool
	}{
		{"nil release", nil, false},
		{"nil info", &release.Release{Version: 1}, false},
		{"pending install", rev(1, release.StatusPendingInstall), true},
		{"pending upgrade", rev(4, release.StatusPendingUpgrade), true},
		{"pending rollback", rev(7, release.StatusPendingRollback), true},
		{"deployed", rev(2, release.StatusDeployed), false},
		{"failed", rev(3, release.StatusFailed), false},
		{"superseded", rev(1, release.StatusSuperseded), false},
		{"uninstalled", rev(5, release.StatusUninstalled), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPending(tt.rel); got != tt.want {
				t.Fatalf("IsPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastGoodRevision(t *testing.T) {
	tests := []struct {
		name    string
		history []*release.Release
		want    int
		wantOK  bool
	}{
		{
			name:    "no history",
			history: nil,
			wantOK:  false,
		},
		{
			// The common pending-install case: one revision that never rolled
			// out. There is nothing to roll back to, so the caller has to
			// uninstall instead.
			name:    "only a pending first install",
			history: []*release.Release{rev(1, release.StatusPendingInstall)},
			wantOK:  false,
		},
		{
			name: "pending upgrade over a deployed revision",
			history: []*release.Release{
				rev(1, release.StatusSuperseded),
				rev(2, release.StatusDeployed),
				rev(3, release.StatusPendingUpgrade),
			},
			want:   2,
			wantOK: true,
		},
		{
			// A failed rollout is not a safe target: it ran and did not work.
			name: "failed revisions are skipped",
			history: []*release.Release{
				rev(1, release.StatusDeployed),
				rev(2, release.StatusFailed),
				rev(3, release.StatusFailed),
				rev(4, release.StatusPendingUpgrade),
			},
			want:   1,
			wantOK: true,
		},
		{
			name: "highest superseded wins when nothing is deployed",
			history: []*release.Release{
				rev(1, release.StatusSuperseded),
				rev(3, release.StatusSuperseded),
				rev(2, release.StatusSuperseded),
				rev(4, release.StatusPendingRollback),
			},
			want:   3,
			wantOK: true,
		},
		{
			name: "every revision pending",
			history: []*release.Release{
				rev(1, release.StatusPendingInstall),
				rev(2, release.StatusPendingUpgrade),
			},
			wantOK: false,
		},
		{
			name:    "malformed revisions are ignored",
			history: []*release.Release{nil, {Version: 9}, rev(2, release.StatusDeployed)},
			want:    2,
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := LastGoodRevision(tt.history)
			if ok != tt.wantOK {
				t.Fatalf("LastGoodRevision() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Fatalf("LastGoodRevision() = %d, want %d", got, tt.want)
			}
		})
	}
}
