package kubernetes_manifest

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestResourceDiff(t *testing.T) {
	h, _ := New(HandlerParams{})

	tests := []struct {
		name     string
		prev     []*kubernetesResource
		curr     []*kubernetesResource
		expected struct {
			apply  []*kubernetesResource
			delete []*kubernetesResource
		}
	}{
		{
			name: "No changes",
			prev: []*kubernetesResource{
				{groupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}, groupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				}, namespace: "default", name: "pod1"},
			},
			curr: []*kubernetesResource{
				{groupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}, groupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				}, namespace: "default", name: "pod1"},
			},
			expected: struct {
				apply  []*kubernetesResource
				delete []*kubernetesResource
			}{
				apply: []*kubernetesResource{
					{groupVersionResource: schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					}, groupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					}, namespace: "default", name: "pod1"},
				},
				delete: []*kubernetesResource{},
			},
		},
		{
			name: "Resource deleted",
			prev: []*kubernetesResource{
				{groupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}, groupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				}, namespace: "default", name: "pod1"},
			},
			curr: []*kubernetesResource{},
			expected: struct {
				apply  []*kubernetesResource
				delete []*kubernetesResource
			}{
				apply: []*kubernetesResource{},
				delete: []*kubernetesResource{
					{groupVersionResource: schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					}, groupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					}, namespace: "default", name: "pod1"},
				},
			},
		},
		{
			name: "Resource added",
			prev: []*kubernetesResource{},
			curr: []*kubernetesResource{
				{groupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}, groupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				}, namespace: "default", name: "pod1"},
			},
			expected: struct {
				apply  []*kubernetesResource
				delete []*kubernetesResource
			}{
				apply: []*kubernetesResource{
					{groupVersionResource: schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					}, groupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					}, namespace: "default", name: "pod1"},
				},
				delete: []*kubernetesResource{},
			},
		},
		{
			name: "Resource modified",
			prev: []*kubernetesResource{
				{groupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}, groupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				}, namespace: "default", name: "pod1"},
			},
			curr: []*kubernetesResource{
				{groupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}, groupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				}, namespace: "default", name: "pod2"},
			},
			expected: struct {
				apply  []*kubernetesResource
				delete []*kubernetesResource
			}{
				apply: []*kubernetesResource{
					{groupVersionResource: schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					}, groupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					}, namespace: "default", name: "pod2"},
				},
				delete: []*kubernetesResource{
					{groupVersionResource: schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "pods",
					}, groupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					}, namespace: "default", name: "pod1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apply, delete := h.resourceDiff(tt.prev, tt.curr)
			if len(apply) != len(tt.expected.apply) {
				t.Errorf("%s : apply mismatch: got %v, want %v", tt.name, apply, tt.expected.apply)
			}
			if len(delete) != len(tt.expected.delete) {
				t.Errorf("%s : delete mismatch: got %v, want %v", tt.name, delete, tt.expected.delete)
			}
		})
	}
}
