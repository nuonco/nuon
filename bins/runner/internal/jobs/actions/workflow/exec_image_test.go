package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func TestActionImageRef(t *testing.T) {
	t.Run("runs the manifest ctl-api pinned", func(t *testing.T) {
		h := &handler{state: &handlerState{plan: &plantypes.ActionWorkflowRunPlan{
			SourceImage:    "curlimages/curl:latest",
			ImageDigestRef: "reg/org/app@sha256:abc",
		}}}

		ref, err := h.actionImageRef()
		require.NoError(t, err)
		assert.Equal(t, "reg/org/app@sha256:abc", ref)
	})

	// Without a digest there is nothing to run but whatever the tag points at
	// now, which is not what Nuon resolved.
	t.Run("fails when the plan carries no digest", func(t *testing.T) {
		h := &handler{state: &handlerState{plan: &plantypes.ActionWorkflowRunPlan{
			SourceImage: "curlimages/curl:latest",
		}}}

		_, err := h.actionImageRef()
		assert.Error(t, err)
	})
}
