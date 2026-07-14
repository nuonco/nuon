package controlplanejob_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	componentactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
)

func TestComponentAndControlPlaneActivitiesRegisterTogether(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()

	require.NotPanics(t, func() {
		env.RegisterActivity(&componentactivities.Activities{})
		acts := &controlplanejob.Activities{}
		for _, act := range acts.AllActivities() {
			env.RegisterActivity(act)
		}
	})
}
