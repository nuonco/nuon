package instances

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/generics"
	instancesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1/instances/v1"
	provisionv1 "github.com/powertoolsdev/mono/pkg/types/workflows/instances/v1"
	workers "github.com/powertoolsdev/mono/services/workers-deployments/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvisionInstances(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*instancesv1.ProvisionRequest]()
	installID := domains.NewInstallID()
	req.InstallIds = []string{installID}
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}

	// register activities
	act := NewActivities(cfg)
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(act.ProvisionInstance, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr *provisionv1.ProvisionRequest) (*provisionv1.ProvisionResponse, error) {
			assert.Nil(t, pr.Validate())
			assert.Equal(t, req.OrgId, pr.OrgId)

			assert.Equal(t, req.InstallIds[0], pr.InstallId)

			return &provisionv1.ProvisionResponse{}, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.ProvisionInstances, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &instancesv1.ProvisionResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
