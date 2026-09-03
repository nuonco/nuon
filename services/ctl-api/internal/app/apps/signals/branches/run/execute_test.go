package run

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

func TestFinalizerPhasesSuccess(t *testing.T) {
	context := &activities.GetPreviewCommentContextOutput{
		Phases: &activities.PRCommentPhases{
			Config:  activities.PRCommentPhaseValidating,
			Builds:  activities.PRCommentPhaseBuilding,
			Install: activities.PRCommentPhaseConfiguring,
		},
	}

	phases := finalizerPhases(context, app.AppBranchRunPreviewModePlanOnly, true)

	require.Equal(t, activities.PRCommentPhaseValid, phases.Config)
	require.Equal(t, activities.PRCommentPhaseValid, phases.Builds)
	require.Equal(t, activities.PRCommentPhaseValid, phases.Install)
}

func TestFinalizerPhasesBuildOnlyOmitsInstall(t *testing.T) {
	context := &activities.GetPreviewCommentContextOutput{
		Phases: &activities.PRCommentPhases{
			Config:  activities.PRCommentPhaseValidating,
			Builds:  activities.PRCommentPhaseBuilding,
			Install: activities.PRCommentPhaseConfiguring,
		},
	}

	phases := finalizerPhases(context, app.AppBranchRunPreviewModeBuildOnly, true)

	require.Equal(t, activities.PRCommentPhaseValid, phases.Config)
	require.Equal(t, activities.PRCommentPhaseValid, phases.Builds)
	require.Empty(t, phases.Install)
}

func TestFinalizerPhasesFailure(t *testing.T) {
	context := &activities.GetPreviewCommentContextOutput{
		Phases: &activities.PRCommentPhases{
			Config:  activities.PRCommentPhaseValid,
			Builds:  activities.PRCommentPhaseBuilding,
			Install: activities.PRCommentPhaseConfiguring,
		},
	}

	phases := finalizerPhases(context, app.AppBranchRunPreviewModeApply, false)

	require.Equal(t, activities.PRCommentPhaseValid, phases.Config)
	require.Equal(t, activities.PRCommentPhaseInvalid, phases.Builds)
	require.Equal(t, activities.PRCommentPhaseInvalid, phases.Install)
}
