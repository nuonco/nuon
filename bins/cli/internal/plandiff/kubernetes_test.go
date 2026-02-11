package plandiff

import (
	"strings"
	"testing"
)

func TestParseKubernetesPlan(t *testing.T) {
	tests := []struct {
		name        string
		plan        *KubernetesPlan
		wantAdd     int
		wantChange  int
		wantDestroy int
		wantErrors  int
	}{
		{
			name: "empty plan",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{},
			},
			wantAdd:     0,
			wantChange:  0,
			wantDestroy: 0,
		},
		{
			name: "add from item type",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Type:      1,
					},
				},
			},
			wantAdd: 1,
		},
		{
			name: "delete from item type",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "ConfigMap",
						Name:      "old-config",
						Namespace: "default",
						Type:      2,
					},
				},
			},
			wantDestroy: 1,
		},
		{
			name: "change from item type",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Type:      3,
					},
				},
			},
			wantChange: 1,
		},
		{
			name: "change from entry type fallback",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Type:      0,
						Entries: []KubernetesDiffEntry{
							{Type: 3, Path: "spec.replicas", Original: "2", Applied: "3"},
						},
					},
				},
			},
			wantChange: 1,
		},
		{
			name: "error items",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "failed-app",
						Namespace: "default",
						Error:     "resource not found",
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "mixed items with errors",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "app1",
						Namespace: "default",
						Type:      1,
					},
					{
						Kind:      "Service",
						Name:      "svc1",
						Namespace: "default",
						Type:      3,
					},
					{
						Kind:      "ConfigMap",
						Name:      "broken",
						Namespace: "default",
						Error:     "validation error",
					},
				},
			},
			wantAdd:    1,
			wantChange: 1,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseKubernetesPlan(tt.plan)

			if parsed.Summary.Add != tt.wantAdd {
				t.Errorf("Summary.Add = %d, want %d", parsed.Summary.Add, tt.wantAdd)
			}
			if parsed.Summary.Change != tt.wantChange {
				t.Errorf("Summary.Change = %d, want %d", parsed.Summary.Change, tt.wantChange)
			}
			if parsed.Summary.Destroy != tt.wantDestroy {
				t.Errorf("Summary.Destroy = %d, want %d", parsed.Summary.Destroy, tt.wantDestroy)
			}
			if len(parsed.Errors) != tt.wantErrors {
				t.Errorf("len(Errors) = %d, want %d", len(parsed.Errors), tt.wantErrors)
			}
		})
	}
}

func TestDetermineKubernetesAction(t *testing.T) {
	tests := []struct {
		name      string
		itemType  int
		entryType int
		before    *string
		after     *string
		want      HelmK8sChangeAction
	}{
		{
			name:     "item type add",
			itemType: 1,
			want:     HelmK8sActionAdded,
		},
		{
			name:     "item type delete",
			itemType: 2,
			want:     HelmK8sActionDestroyed,
		},
		{
			name:     "item type change",
			itemType: 3,
			want:     HelmK8sActionChanged,
		},
		{
			name:      "fallback to entry type add",
			itemType:  0,
			entryType: 1,
			want:      HelmK8sActionAdded,
		},
		{
			name:      "fallback to entry type delete",
			itemType:  0,
			entryType: 2,
			want:      HelmK8sActionDestroyed,
		},
		{
			name:      "fallback to entry type change",
			itemType:  0,
			entryType: 3,
			want:      HelmK8sActionChanged,
		},
		{
			name:      "infer add from after only",
			itemType:  0,
			entryType: 0,
			after:     strPtr("new value"),
			want:      HelmK8sActionAdded,
		},
		{
			name:      "infer delete from before only",
			itemType:  0,
			entryType: 0,
			before:    strPtr("old value"),
			want:      HelmK8sActionDestroyed,
		},
		{
			name:      "infer change from both",
			itemType:  0,
			entryType: 0,
			before:    strPtr("old"),
			after:     strPtr("new"),
			want:      HelmK8sActionChanged,
		},
		{
			name:      "default to change when no info",
			itemType:  0,
			entryType: 0,
			want:      HelmK8sActionChanged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineKubernetesAction(tt.itemType, tt.entryType, tt.before, tt.after)
			if got != tt.want {
				t.Errorf("determineKubernetesAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatKubernetesPlan(t *testing.T) {
	tests := []struct {
		name     string
		plan     *KubernetesPlan
		contains []string
	}{
		{
			name: "no changes",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{},
			},
			contains: []string{"No changes"},
		},
		{
			name: "with plan text",
			plan: &KubernetesPlan{
				Plan:           "apply manifests",
				K8sContentDiff: []KubernetesDiffItem{},
			},
			contains: []string{"Plan:", "apply manifests"},
		},
		{
			name: "add action",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "default",
						Type:      1,
					},
				},
			},
			contains: []string{"1 to add", "Deployment", "test-app"},
		},
		{
			name: "with api version",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						API:       "apps/v1",
						Kind:      "Deployment",
						Name:      "test-app",
						Namespace: "prod",
						Type:      1,
					},
				},
			},
			contains: []string{"apps/v1"},
		},
		{
			name: "with namespace",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Service",
						Name:      "my-svc",
						Namespace: "kube-system",
						Type:      1,
					},
				},
			},
			contains: []string{"kube-system/my-svc"},
		},
		{
			name: "with errors",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{
						Kind:      "Deployment",
						Name:      "broken",
						Namespace: "default",
						Error:     "validation failed: missing required field",
					},
				},
			},
			contains: []string{"Errors", "broken", "validation failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseKubernetesPlan(tt.plan)
			output := FormatKubernetesPlan(parsed, tt.plan.Plan)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatKubernetesPlan output missing %q\nOutput: %s", want, output)
				}
			}
		})
	}
}

func TestHasKubernetesChanges(t *testing.T) {
	tests := []struct {
		name       string
		plan       *KubernetesPlan
		wantChange bool
	}{
		{
			name: "no changes",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{},
			},
			wantChange: false,
		},
		{
			name: "has add",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{Kind: "Deployment", Name: "app", Type: 1},
				},
			},
			wantChange: true,
		},
		{
			name: "has change",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{Kind: "Deployment", Name: "app", Type: 3},
				},
			},
			wantChange: true,
		},
		{
			name: "has destroy",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{Kind: "ConfigMap", Name: "old", Type: 2},
				},
			},
			wantChange: true,
		},
		{
			name: "has errors only",
			plan: &KubernetesPlan{
				K8sContentDiff: []KubernetesDiffItem{
					{Kind: "Deployment", Name: "broken", Error: "failed"},
				},
			},
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseKubernetesPlan(tt.plan)
			got := HasKubernetesChanges(parsed)

			if got != tt.wantChange {
				t.Errorf("HasKubernetesChanges() = %v, want %v", got, tt.wantChange)
			}
		})
	}
}

func TestIncrementKubernetesSummary(t *testing.T) {
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
			incrementKubernetesSummary(s, tt.action)

			if got := tt.checkFn(s); got != tt.wantVal {
				t.Errorf("incrementKubernetesSummary(%s) counter = %d, want %d", tt.action, got, tt.wantVal)
			}
		})
	}
}
