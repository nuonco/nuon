package approvalplan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelmApprovalPlan_IsNoop(t *testing.T) {
	tests := []struct {
		name     string
		planJSON string
		want     bool
	}{
		{
			name:     "no content diff key",
			planJSON: `{}`,
			want:     false,
		},
		{
			name:     "empty content diff",
			planJSON: `{"helm_content_diff": []}`,
			want:     false,
		},
		{
			name:     "content diff present",
			planJSON: `{"helm_content_diff": [{"name": "cm"}]}`,
			want:     false,
		},
		{
			name:     "empty diff against a deployed release",
			planJSON: `{"helm_content_diff": [], "helm_release_status": "deployed"}`,
			want:     true,
		},
		// a timed-out attempt leaves the new manifest on a pending release
		{
			name:     "empty diff against a pending-install release",
			planJSON: `{"helm_content_diff": [], "helm_release_status": "pending-install"}`,
			want:     false,
		},
		{
			name:     "empty diff against a pending-upgrade release",
			planJSON: `{"helm_content_diff": [], "helm_release_status": "pending-upgrade"}`,
			want:     false,
		},
		{
			name:     "empty diff against a failed release",
			planJSON: `{"helm_content_diff": [], "helm_release_status": "failed"}`,
			want:     false,
		},
		// a missing status means the release state is unknown not noop
		{
			name:     "empty diff with an empty status",
			planJSON: `{"helm_content_diff": [], "helm_release_status": ""}`,
			want:     false,
		},
		{
			name:     "uninstall with nothing to remove",
			planJSON: `{"op": "uninstall", "helm_content_diff": []}`,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHelmApprovalPlen([]byte(tt.planJSON)).IsNoop()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
