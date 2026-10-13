package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/ecrrepository"
)

func TestDeprovisionECRRepositoryRoutesByCloud(t *testing.T) {
	deleteActivities := []string{"DeleteECRRepository", "DeleteGARPackage", "DeleteACRRepository"}

	tests := map[string]struct {
		cloudProvider string
		deleteErr     error
		wantActivity  string
	}{
		"aws default":        {cloudProvider: "", wantActivity: "DeleteECRRepository"},
		"aws":                {cloudProvider: "aws", wantActivity: "DeleteECRRepository"},
		"gcp":                {cloudProvider: "gcp", wantActivity: "DeleteGARPackage"},
		"azure":              {cloudProvider: "azure", wantActivity: "DeleteACRRepository"},
		"aws delete fails":   {cloudProvider: "aws", deleteErr: errors.New("boom"), wantActivity: "DeleteECRRepository"},
		"gcp delete fails":   {cloudProvider: "gcp", deleteErr: errors.New("boom"), wantActivity: "DeleteGARPackage"},
		"azure delete fails": {cloudProvider: "azure", deleteErr: errors.New("boom"), wantActivity: "DeleteACRRepository"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ts testsuite.WorkflowTestSuite
			env := ts.NewTestWorkflowEnvironment()

			started := map[string]int{}
			env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, _ converter.EncodedValues) {
				started[info.ActivityType.Name]++
			})

			w := &Workflows{cfg: &internal.Config{CloudProvider: tt.cloudProvider}}
			for _, fn := range w.All() {
				env.RegisterWorkflow(fn)
			}

			env.OnActivity((*ecrrepository.Activities).DeleteECRRepository, mock.Anything, mock.Anything).
				Return(&ecrrepository.DeleteECRRepositoryResponse{}, tt.deleteErr).Maybe()
			env.OnActivity((*ecrrepository.Activities).DeleteGARPackage, mock.Anything, mock.Anything).
				Return(&ecrrepository.DeleteGARPackageResponse{}, tt.deleteErr).Maybe()
			env.OnActivity((*ecrrepository.Activities).DeleteACRRepository, mock.Anything, mock.Anything).
				Return(&ecrrepository.DeleteACRRepositoryResponse{}, tt.deleteErr).Maybe()

			env.ExecuteWorkflow("DeprovisionECRRepository", &ecrrepository.DeprovisionECRRepositoryRequest{
				OrgID: "orgabc",
				AppID: "appdef",
			})

			require.True(t, env.IsWorkflowCompleted())
			require.NoError(t, env.GetWorkflowError())

			for _, act := range deleteActivities {
				if act == tt.wantActivity {
					require.Positive(t, started[act], "expected %s to run", act)
					continue
				}
				require.Zero(t, started[act], "expected %s not to run", act)
			}
		})
	}
}
