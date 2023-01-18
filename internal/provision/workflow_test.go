package provision

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/powertoolsdev/go-generics"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
)

func TestProvision(t *testing.T) {
	wf := NewWorkflow(workers.Config{
		Bucket:                       "nuon-installations",
		WaypointTokenSecretNamespace: "default",
		WaypointServerRootDomain:     "test.nuon.co",
	})
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities(nil)

	req := generics.GetFakeObj[*instancesv1.ProvisionRequest]()

	env.OnActivity(a.SendHostnameNotification, mock.Anything, mock.Anything).
		Return(func(_ context.Context, shnReq SendHostnameNotificationRequest) (SendHostnameNotificationResponse, error) {
			var resp SendHostnameNotificationResponse
			assert.Nil(t, shnReq.validate())

			return resp, nil
		})

	env.ExecuteWorkflow(wf.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	resp := &instancesv1.ProvisionResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
