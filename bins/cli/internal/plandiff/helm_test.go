package plandiff

import (
	"strings"
	"testing"
)

func TestParseHelmPlan(t *testing.T) {
	tests := []struct {
		name        string
		plan        *HelmPlan
		wantAdd     int
		wantChange  int
		wantDestroy int
	}{
		{
			name: "empty plan",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{},
			},
			wantAdd:     0,
			wantChange:  0,
			wantDestroy: 0,
		},
		{
			name: "single add from entry type",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Entries: []HelmDiffEntry{
							{Type: 1, Path: "spec.replicas", Applied: "3"},
						},
					},
				},
			},
			wantAdd: 1,
		},
		{
			name: "single delete from entry type",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "ConfigMap",
						Name:      "old-config",
						Namespace: "default",
						Entries: []HelmDiffEntry{
							{Type: 2, Path: "data.key", Original: "value"},
						},
					},
				},
			},
			wantDestroy: 1,
		},
		{
			name: "change from entry type",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Entries: []HelmDiffEntry{
							{Type: 3, Path: "spec.replicas", Original: "2", Applied: "3"},
						},
					},
				},
			},
			wantChange: 1,
		},
		{
			name: "inferred add from before/after",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Service",
						Name:      "new-svc",
						Namespace: "default",
						After:     "port: 80",
					},
				},
			},
			wantAdd: 1,
		},
		{
			name: "inferred delete from before/after",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Service",
						Name:      "old-svc",
						Namespace: "default",
						Before:    "port: 80",
					},
				},
			},
			wantDestroy: 1,
		},
		{
			name: "inferred change from before/after",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Service",
						Name:      "svc",
						Namespace: "default",
						Before:    "port: 80",
						After:     "port: 8080",
					},
				},
			},
			wantChange: 1,
		},
		{
			name: "multiple items",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Deployment",
						Name:      "app1",
						Namespace: "default",
						Entries:   []HelmDiffEntry{{Type: 1, Path: "spec.replicas", Applied: "3"}},
					},
					{
						Kind:      "Service",
						Name:      "svc1",
						Namespace: "default",
						Entries:   []HelmDiffEntry{{Type: 3, Path: "spec.port", Original: "80", Applied: "8080"}},
					},
					{
						Kind:      "ConfigMap",
						Name:      "cm1",
						Namespace: "default",
						Entries:   []HelmDiffEntry{{Type: 2, Path: "data.key", Original: "value"}},
					},
				},
			},
			wantAdd:     1,
			wantChange:  1,
			wantDestroy: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseHelmPlan(tt.plan)

			if parsed.Summary.Add != tt.wantAdd {
				t.Errorf("Summary.Add = %d, want %d", parsed.Summary.Add, tt.wantAdd)
			}
			if parsed.Summary.Change != tt.wantChange {
				t.Errorf("Summary.Change = %d, want %d", parsed.Summary.Change, tt.wantChange)
			}
			if parsed.Summary.Destroy != tt.wantDestroy {
				t.Errorf("Summary.Destroy = %d, want %d", parsed.Summary.Destroy, tt.wantDestroy)
			}
		})
	}
}

func TestDetermineHelmAction(t *testing.T) {
	tests := []struct {
		name      string
		entryType int
		before    *string
		after     *string
		want      HelmK8sChangeAction
	}{
		{
			name:      "entry type add",
			entryType: 1,
			want:      HelmK8sActionAdded,
		},
		{
			name:      "entry type delete",
			entryType: 2,
			want:      HelmK8sActionDestroyed,
		},
		{
			name:      "entry type change",
			entryType: 3,
			want:      HelmK8sActionChanged,
		},
		{
			name:      "infer add from after only",
			entryType: 0,
			after:     strPtr("new value"),
			want:      HelmK8sActionAdded,
		},
		{
			name:      "infer delete from before only",
			entryType: 0,
			before:    strPtr("old value"),
			want:      HelmK8sActionDestroyed,
		},
		{
			name:      "infer change from both",
			entryType: 0,
			before:    strPtr("old"),
			after:     strPtr("new"),
			want:      HelmK8sActionChanged,
		},
		{
			name:      "default to change when no info",
			entryType: 0,
			want:      HelmK8sActionChanged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineHelmAction(tt.entryType, tt.before, tt.after)
			if got != tt.want {
				t.Errorf("determineHelmAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatHelmPlan(t *testing.T) {
	tests := []struct {
		name     string
		plan     *HelmPlan
		contains []string
	}{
		{
			name: "no changes",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{},
			},
			contains: []string{"No changes"},
		},
		{
			name: "add action",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Entries: []HelmDiffEntry{
							{Type: 1, Path: "spec.replicas", Applied: "3"},
						},
					},
				},
			},
			contains: []string{"1 to add", "Deployment", "test-app"},
		},
		{
			name: "with api version",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						API:       "apps/v1",
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "prod",
						Entries: []HelmDiffEntry{
							{Type: 1, Path: "spec.replicas", Applied: "3"},
						},
					},
				},
			},
			contains: []string{"apps/v1"},
		},
		{
			name: "with namespace",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind:      "Service",
						Name:      "my-svc",
						Namespace: "kube-system",
						After:     "port: 443",
					},
				},
			},
			contains: []string{"kube-system/my-svc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseHelmPlan(tt.plan)
			output := FormatHelmPlan(parsed)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatHelmPlan output missing %q\nOutput: %s", want, output)
				}
			}
		})
	}
}

func TestHasHelmChanges(t *testing.T) {
	tests := []struct {
		name       string
		plan       *HelmPlan
		wantChange bool
	}{
		{
			name: "no changes",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{},
			},
			wantChange: false,
		},
		{
			name: "has add",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{Kind: "Deployment", Name: "app", After: "replicas: 1"},
				},
			},
			wantChange: true,
		},
		{
			name: "has change",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{
						Kind: "Deployment",
						Name: "app",
						Entries: []HelmDiffEntry{
							{Type: 3, Path: "spec.replicas", Original: "1", Applied: "2"},
						},
					},
				},
			},
			wantChange: true,
		},
		{
			name: "has destroy",
			plan: &HelmPlan{
				HelmContentDiff: []HelmDiffItem{
					{Kind: "ConfigMap", Name: "old", Before: "key: value"},
				},
			},
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseHelmPlan(tt.plan)
			got := HasHelmChanges(parsed)

			if got != tt.wantChange {
				t.Errorf("HasHelmChanges() = %v, want %v", got, tt.wantChange)
			}
		})
	}
}

func TestIncrementHelmSummary(t *testing.T) {
	tests := []struct {
		action  HelmK8sChangeAction
		checkFn func(s *Summary) int
		wantVal int
	}{
		{HelmK8sActionAdd, func(s *Summary) int { return s.Add }, 1},
		{HelmK8sActionAdded, func(s *Summary) int { return s.Add }, 1},
		{HelmK8sActionChange, func(s *Summary) int { return s.Change }, 1},
		{HelmK8sActionChanged, func(s *Summary) int { return s.Change }, 1},
		{HelmK8sActionDestroy, func(s *Summary) int { return s.Destroy }, 1},
		{HelmK8sActionDestroyed, func(s *Summary) int { return s.Destroy }, 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			s := &Summary{}
			incrementHelmSummary(s, tt.action)

			if got := tt.checkFn(s); got != tt.wantVal {
				t.Errorf("incrementHelmSummary(%s) counter = %d, want %d", tt.action, got, tt.wantVal)
			}
		})
	}
}
