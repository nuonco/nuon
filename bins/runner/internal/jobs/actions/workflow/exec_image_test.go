package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func TestPullSourceImageDirectly(t *testing.T) {
	tests := []struct {
		name   string
		plan   *plantypes.ActionWorkflowRunPlan
		devEnv bool
		want   bool
	}{
		{
			name:   "mirrored plan in dev pulls the public source directly",
			plan:   &plantypes.ActionWorkflowRunPlan{SourceImage: "curlimages/curl:latest", ImageDigestRef: "reg/org/app@sha256:abc"},
			devEnv: true,
			want:   true,
		},
		{
			name:   "install-registry plan never takes the dev shortcut",
			plan:   &plantypes.ActionWorkflowRunPlan{SourceImage: "reg/org/app/tools@sha256:abc", ImageDigestRef: "reg/org/app/tools@sha256:abc"},
			devEnv: true,
			want:   false,
		},
		{
			name:   "mirrored plan outside dev uses the mirror",
			plan:   &plantypes.ActionWorkflowRunPlan{SourceImage: "curlimages/curl:latest", ImageDigestRef: "reg/org/app@sha256:abc"},
			devEnv: false,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.devEnv {
				t.Setenv("NUON_DEV_REAL_IMAGE_ACTIONS", "true")
				t.Setenv("ENV", "development")
			}

			h := &handler{state: &handlerState{plan: tt.plan}}
			assert.Equal(t, tt.want, h.pullSourceImageDirectly())
		})
	}
}
