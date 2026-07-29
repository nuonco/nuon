package runnerunhealthy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSignalLifecycleContext(t *testing.T) {
	sig := &Signal{
		RunnerID:             "run_1",
		RunnerName:           "Default runner",
		OrgID:                "org_1",
		OrgName:              "Acme",
		FromStatus:           app.RunnerStatusActive,
		ToStatus:             app.RunnerStatusOffline,
		Reason:               "no active install process",
		RunnerGroupID:        "rug_1",
		RunnerGroupType:      app.RunnerGroupTypeInstall,
		RunnerGroupOwnerID:   "ins_1",
		RunnerGroupOwnerType: "installs",
		RunnerGroupOwnerName: "Production",
	}

	require.NoError(t, sig.Validate(nil))
	ctx := sig.LifecycleContext()
	require.NotNil(t, ctx.InstallID)
	assert.Equal(t, "ins_1", *ctx.InstallID)
	assert.Equal(t, "ins_1", ctx.OwnerID)
	assert.Equal(t, "installs", ctx.OwnerType)
	assert.Equal(t, "Production", ctx.OwnerName)
	assert.Equal(t, "run_1", ctx.Metadata["runner_id"])
	assert.Equal(t, "active", ctx.Metadata["from_status"])
	assert.Equal(t, "offline", ctx.Metadata["to_status"])
}

func TestSignalRejectsNonUnhealthyTransition(t *testing.T) {
	sig := &Signal{
		RunnerID:             "run_1",
		OrgID:                "org_1",
		FromStatus:           app.RunnerStatusOffline,
		ToStatus:             app.RunnerStatusActive,
		Reason:               "runner healthy",
		RunnerGroupID:        "rug_1",
		RunnerGroupType:      app.RunnerGroupTypeOrg,
		RunnerGroupOwnerID:   "org_1",
		RunnerGroupOwnerType: "orgs",
	}

	require.Error(t, sig.Validate(nil))
}
