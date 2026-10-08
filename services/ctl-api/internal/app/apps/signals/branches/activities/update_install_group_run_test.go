package activities

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installgrouprelease"
)

func TestPreserveSupersededInstalls(t *testing.T) {
	stored := []app.InstallGroupRunInstall{{
		InstallID:         "ins_1",
		Status:            installgrouprelease.StatusSuperseded,
		SupersededByRunID: "run_2",
	}, {
		InstallID: "ins_2",
		Status:    "in-progress",
	}}
	incoming := []app.InstallGroupRunInstall{{
		InstallID: "ins_1",
		Status:    "cancelled",
	}, {
		InstallID: "ins_2",
		Status:    "success",
	}}

	got := preserveSupersededInstalls(stored, incoming)
	require.Equal(t, installgrouprelease.StatusSuperseded, got[0].Status)
	require.Equal(t, "run_2", got[0].SupersededByRunID)
	require.Equal(t, "success", got[1].Status)
	require.Equal(t, "cancelled", incoming[0].Status)
}
