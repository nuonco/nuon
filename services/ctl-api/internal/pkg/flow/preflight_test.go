package flow

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestWorkflowPreflightResultBlocked(t *testing.T) {
	tests := map[string]struct {
		severity compositeerrors.Severity
		blocked  bool
	}{
		"info":    {severity: compositeerrors.SeverityInfo},
		"warning": {severity: compositeerrors.SeverityWarning},
		"error":   {severity: compositeerrors.SeverityError, blocked: true},
		"fatal":   {severity: compositeerrors.SeverityFatal, blocked: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := &WorkflowPreflightResult{Findings: []*compositeerrors.CompositeErrorData{
				{Severity: test.severity},
			}}
			require.Equal(t, test.blocked, result.Blocked())
		})
	}

	var result *WorkflowPreflightResult
	require.False(t, result.Blocked())
}
