package features

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSyncControlPlaneBuildsAndOrgRunner(t *testing.T) {
	cpKey := string(app.OrgFeatureControlPlaneBuilds)
	runnerKey := string(app.OrgFeatureOrgRunner)

	tests := []struct {
		name       string
		current    map[string]bool
		updated    map[string]bool
		wantCP     bool
		wantRunner bool
		wantErr    bool
	}{
		{
			name:       "disabling control-plane-builds re-enables org-runner",
			current:    map[string]bool{cpKey: true, runnerKey: false},
			updated:    map[string]bool{cpKey: false, runnerKey: false},
			wantCP:     false,
			wantRunner: true,
		},
		{
			name:       "enabling control-plane-builds disables org-runner",
			current:    map[string]bool{cpKey: false, runnerKey: true},
			updated:    map[string]bool{cpKey: true, runnerKey: true},
			wantCP:     true,
			wantRunner: false,
		},
		{
			name:       "disabling org-runner enables control-plane-builds",
			current:    map[string]bool{cpKey: false, runnerKey: true},
			updated:    map[string]bool{cpKey: false, runnerKey: false},
			wantCP:     true,
			wantRunner: false,
		},
		{
			name:       "enabling org-runner disables control-plane-builds",
			current:    map[string]bool{cpKey: true, runnerKey: false},
			updated:    map[string]bool{cpKey: true, runnerKey: true},
			wantCP:     false,
			wantRunner: true,
		},
		{
			name:       "neither flag changed leaves both untouched",
			current:    map[string]bool{cpKey: true, runnerKey: false},
			updated:    map[string]bool{cpKey: true, runnerKey: false},
			wantCP:     true,
			wantRunner: false,
		},
		{
			name:       "consistent simultaneous change is accepted",
			current:    map[string]bool{cpKey: true, runnerKey: false},
			updated:    map[string]bool{cpKey: false, runnerKey: true},
			wantCP:     false,
			wantRunner: true,
		},
		{
			name:    "conflicting simultaneous change (both true) errors",
			current: map[string]bool{cpKey: false, runnerKey: false},
			updated: map[string]bool{cpKey: true, runnerKey: true},
			wantErr: true,
		},
		{
			name:    "conflicting simultaneous change (both false) errors",
			current: map[string]bool{cpKey: true, runnerKey: true},
			updated: map[string]bool{cpKey: false, runnerKey: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := syncControlPlaneBuildsAndOrgRunner(tt.current, tt.updated)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.updated[cpKey] != tt.wantCP {
				t.Errorf("control-plane-builds = %v, want %v", tt.updated[cpKey], tt.wantCP)
			}
			if tt.updated[runnerKey] != tt.wantRunner {
				t.Errorf("org-runner = %v, want %v", tt.updated[runnerKey], tt.wantRunner)
			}
		})
	}
}
