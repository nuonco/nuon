package branchrunerrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestConfigValidationFailedError(t *testing.T) {
	err := &ConfigValidationFailedError{Detail: "stack name references sandbox outputs"}
	data, buildErr := compositeerrors.New(err, compositeerrors.WithSource("app_branch_runs", "arn-example"))
	require.NoError(t, buildErr)
	require.Equal(t, ConfigValidationFailedType, data.Type)
	require.Equal(t, compositeerrors.SeverityFatal, data.Severity)
	require.True(t, data.Hints.Terminal())
	require.Len(t, data.Sections, 3)
	require.Equal(t, compositeerrors.SectionCode, data.Sections[1].Kind)
}

func TestValidationDetail(t *testing.T) {
	cause := temporal.NewNonRetryableApplicationError(
		"invalid template references",
		ConfigValidationFailedTemporalType,
		nil,
	)
	detail, ok := ValidationDetail(fmt.Errorf("activity failed: %w", cause))
	require.True(t, ok)
	require.Equal(t, "invalid template references", detail)

	_, ok = ValidationDetail(temporal.NewApplicationError("database unavailable", "unavailable"))
	require.False(t, ok)
}
