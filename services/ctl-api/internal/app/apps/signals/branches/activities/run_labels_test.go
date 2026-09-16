package activities

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildRunLabelsDraftMode(t *testing.T) {
	labels := BuildRunLabels(&TriggerAppBranchRunFromVCSPushRequest{
		Draft: true,
	})

	require.Equal(t, "true", labels[app.AppBranchRunLabelIsDraftMode])
}

func TestBuildRunLabelsOmitsDraftModeForReadyPR(t *testing.T) {
	labels := BuildRunLabels(&TriggerAppBranchRunFromVCSPushRequest{})

	_, ok := labels[app.AppBranchRunLabelIsDraftMode]
	require.False(t, ok)
}
