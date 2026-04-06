package helpers

import (
	"testing"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSandboxDependentComponents(t *testing.T) {
	tests := []struct {
		name        string
		connections []app.ComponentConfigConnection
		wantIDs     map[string]bool
	}{
		{
			name:        "no components",
			connections: nil,
			wantIDs:     map[string]bool{},
		},
		{
			name: "no sandbox refs",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "a", Refs: []refs.Ref{{Type: refs.RefTypeInputs}}},
				{ComponentID: "b", Refs: []refs.Ref{{Type: refs.RefTypeComponents}}},
			},
			wantIDs: map[string]bool{},
		},
		{
			name: "direct sandbox dependent only",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "a", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "b", Refs: []refs.Ref{{Type: refs.RefTypeInputs}}},
			},
			wantIDs: map[string]bool{"a": true},
		},
		{
			name: "transitive dependent via component dependency",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "sandbox-dep", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "child", ComponentDependencyIDs: pq.StringArray{"sandbox-dep"}},
			},
			wantIDs: map[string]bool{"sandbox-dep": true, "child": true},
		},
		{
			name: "deep transitive chain",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "a", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "b", ComponentDependencyIDs: pq.StringArray{"a"}},
				{ComponentID: "c", ComponentDependencyIDs: pq.StringArray{"b"}},
				{ComponentID: "d", ComponentDependencyIDs: pq.StringArray{"c"}},
			},
			wantIDs: map[string]bool{"a": true, "b": true, "c": true, "d": true},
		},
		{
			name: "unrelated components excluded from transitive chain",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "a", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "b", ComponentDependencyIDs: pq.StringArray{"a"}},
				{ComponentID: "c", Refs: []refs.Ref{{Type: refs.RefTypeInputs}}},
				{ComponentID: "d", ComponentDependencyIDs: pq.StringArray{"c"}},
			},
			wantIDs: map[string]bool{"a": true, "b": true},
		},
		{
			name: "diamond dependency",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "root", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "left", ComponentDependencyIDs: pq.StringArray{"root"}},
				{ComponentID: "right", ComponentDependencyIDs: pq.StringArray{"root"}},
				{ComponentID: "bottom", ComponentDependencyIDs: pq.StringArray{"left", "right"}},
			},
			wantIDs: map[string]bool{"root": true, "left": true, "right": true, "bottom": true},
		},
		{
			name: "multiple sandbox roots",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "s1", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "s2", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "child-of-s1", ComponentDependencyIDs: pq.StringArray{"s1"}},
				{ComponentID: "child-of-s2", ComponentDependencyIDs: pq.StringArray{"s2"}},
				{ComponentID: "independent"},
			},
			wantIDs: map[string]bool{"s1": true, "s2": true, "child-of-s1": true, "child-of-s2": true},
		},
		{
			name: "component with mixed refs includes sandbox",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "a", Refs: []refs.Ref{
					{Type: refs.RefTypeInputs},
					{Type: refs.RefTypeSandbox},
					{Type: refs.RefTypeComponents},
				}},
			},
			wantIDs: map[string]bool{"a": true},
		},
		{
			name: "partial chain with independent branch",
			connections: []app.ComponentConfigConnection{
				{ComponentID: "a", Refs: []refs.Ref{{Type: refs.RefTypeSandbox}}},
				{ComponentID: "b", ComponentDependencyIDs: pq.StringArray{"a"}},
				{ComponentID: "c", ComponentDependencyIDs: pq.StringArray{"a", "x"}},
				{ComponentID: "x", Refs: []refs.Ref{{Type: refs.RefTypeInputs}}},
			},
			wantIDs: map[string]bool{"a": true, "b": true, "c": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sandboxDependentComponents(tt.connections)

			if len(got) != len(tt.wantIDs) {
				t.Fatalf("got %d components %v, want %d components %v", len(got), got, len(tt.wantIDs), tt.wantIDs)
			}
			for id := range tt.wantIDs {
				if !got[id] {
					t.Errorf("expected component %q to be sandbox-dependent, but it was not", id)
				}
			}
			for id := range got {
				if !tt.wantIDs[id] {
					t.Errorf("unexpected sandbox-dependent component %q", id)
				}
			}
		})
	}
}
