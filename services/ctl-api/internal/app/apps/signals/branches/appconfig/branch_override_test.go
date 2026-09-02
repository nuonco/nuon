package appconfig

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPreviewBranchName(t *testing.T) {
	require.Equal(t, "main", previewBranchName("main", nil))
	require.Equal(t, "main", previewBranchName("main", &app.AppBranchRunPreview{}))
	require.Equal(t, "feature/payments", previewBranchName("main", &app.AppBranchRunPreview{
		GitRef: "feature/payments",
	}))
}
