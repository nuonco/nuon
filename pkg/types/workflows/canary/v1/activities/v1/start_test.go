package activitiesv1

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/stretchr/testify/assert"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func Test_StartWorkflowRequest(t *testing.T) {
	t.Run("test that a valid workflow run", func(t *testing.T) {
		req := &canaryv1.ProvisionRequest{}
		any, err := anypb.New(req)
		assert.NoError(t, err)

		obj := &StartWorkflowRequest{
			Namespace:    "orgs",
			WorkflowName: "Provision",
			Request:      any,
		}
		err = obj.Validate()
		assert.NoError(t, err)
	})
}

func Test_StartWorkflowResponse(t *testing.T) {
	t.Run("test workflow with valid id", func(t *testing.T) {
		obj := &StartWorkflowResponse{
			WorkflowId: domains.NewCanaryID(),
		}
		err := obj.Validate()
		assert.NoError(t, err)
	})

	t.Run("test workflow without id", func(t *testing.T) {
		obj := &StartWorkflowResponse{
			WorkflowId: "",
		}
		err := obj.Validate()
		assert.Error(t, err)
	})
}
