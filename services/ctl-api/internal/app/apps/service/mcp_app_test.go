package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestMCPAppWithDerivedStatusUsesStatusV2(t *testing.T) {
	input := &app.App{
		Status:            app.AppStatusActive,
		StatusDescription: "legacy description",
		StatusV2: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: "artifact registry failed",
		},
	}

	result := mcpAppWithDerivedStatus(input)

	require.Equal(t, app.AppStatusError, result.Status)
	require.Equal(t, "artifact registry failed", result.StatusDescription)
	require.Equal(t, app.AppStatusActive, input.Status)
	require.Equal(t, "legacy description", input.StatusDescription)
}

func TestMCPAppWithDerivedStatusFallsBackToLegacyStatus(t *testing.T) {
	input := &app.App{
		Status:            app.AppStatusActive,
		StatusDescription: "legacy description",
	}

	result := mcpAppWithDerivedStatus(input)

	require.Equal(t, input.Status, result.Status)
	require.Equal(t, input.StatusDescription, result.StatusDescription)
}
