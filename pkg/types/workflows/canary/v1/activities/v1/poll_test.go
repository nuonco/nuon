package activitiesv1

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/stretchr/testify/assert"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func Test_PollWorkflowRequest(t *testing.T) {
	t.Run("test a poll workflow request", func(t *testing.T) {
		req := &PollWorkflowRequest{
			Namespace:    "orgs",
			WorkflowName: "Provision",
			WorkflowId:   domains.NewCanaryID(),
		}
		err := req.Validate()
		assert.NoError(t, err)
	})
}

func Test_PollWorkflowResponse(t *testing.T) {
	t.Run("a valid poll workflow response", func(t *testing.T) {
		resp := &canaryv1.ProvisionResponse{}
		any, err := anypb.New(resp)
		assert.NoError(t, err)

		req := &PollWorkflowResponse{
			Step:     &canaryv1.Step{},
			Response: any,
		}
		err = req.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid poll workflow response", func(t *testing.T) {
		resp := &canaryv1.ProvisionResponse{}
		any, err := anypb.New(resp)
		assert.NoError(t, err)

		req := &PollWorkflowResponse{
			Response: any,
		}
		err = req.Validate()
		assert.Error(t, err)
	})
}
