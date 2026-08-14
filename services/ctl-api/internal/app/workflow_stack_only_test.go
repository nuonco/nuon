package app

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
)

func TestWorkflowIsStackOnly(t *testing.T) {
	for name, tc := range map[string]struct {
		metadata pgtype.Hstore
		want     bool
	}{
		"set":       {pgtype.Hstore{WorkflowMetadataKeyStackOnly: generics.ToPtr("true")}, true},
		"false":     {pgtype.Hstore{WorkflowMetadataKeyStackOnly: generics.ToPtr("false")}, false},
		"absent":    {pgtype.Hstore{}, false},
		"nil":       {nil, false},
		"nil value": {pgtype.Hstore{WorkflowMetadataKeyStackOnly: nil}, false},
	} {
		t.Run(name, func(t *testing.T) {
			wf := Workflow{Metadata: tc.metadata}
			if got := wf.IsStackOnly(); got != tc.want {
				t.Fatalf("IsStackOnly() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWorkflowIsInputsOnly(t *testing.T) {
	for name, tc := range map[string]struct {
		metadata pgtype.Hstore
		want     bool
	}{
		"set":    {pgtype.Hstore{WorkflowMetadataKeyInputsOnly: generics.ToPtr("true")}, true},
		"false":  {pgtype.Hstore{WorkflowMetadataKeyInputsOnly: generics.ToPtr("false")}, false},
		"absent": {pgtype.Hstore{}, false},
	} {
		t.Run(name, func(t *testing.T) {
			wf := Workflow{Metadata: tc.metadata}
			if got := wf.IsInputsOnly(); got != tc.want {
				t.Fatalf("IsInputsOnly() = %v, want %v", got, tc.want)
			}
		})
	}
}
