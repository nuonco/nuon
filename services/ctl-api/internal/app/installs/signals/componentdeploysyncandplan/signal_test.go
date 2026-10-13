package componentdeploysyncandplan

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type SignalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestSignalTestSuite(t *testing.T) {
	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) TestExecute() {
	// TODO: implement test
	s.T().Skip("not yet implemented")
}

func (s *SignalTestSuite) cloneBuildID(sig *Signal) (string, []string) {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity((*activities.Activities).GetDeploy, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.InstallDeploy{ID: "dpl-requested", ComponentBuildID: "bld-requested"}, nil)

	var lookedUp []string
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args converter.EncodedValues) {
		if info.ActivityType.Name != "GetDeploy" {
			return
		}
		var req activities.GetDeployRequest
		if err := args.Get(&req); err == nil {
			lookedUp = append(lookedUp, req.DeployID)
		}
	})

	var defs []signal.CloneStepDef
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		var err error
		defs, err = sig.Clone(ctx, "sync and plan whoami")
		return err
	})
	require.NoError(s.T(), env.GetWorkflowError())
	require.Len(s.T(), defs, 1)

	cloned, ok := defs[0].Signal.(*Signal)
	require.True(s.T(), ok)
	require.Empty(s.T(), cloned.DeployID, "a retry must create a fresh deploy")
	return cloned.BuildID, lookedUp
}

func (s *SignalTestSuite) TestCloneKeepsManualDeployBuild() {
	got, lookedUp := s.cloneBuildID(&Signal{
		InstallID:   "inl-1",
		ComponentID: "cmp-1",
		DeployID:    "dpl-requested",
	})
	require.Equal(s.T(), "bld-requested", got)
	require.Equal(s.T(), []string{"dpl-requested"}, lookedUp)
}

func (s *SignalTestSuite) TestCloneKeepsPinnedBuild() {
	got, lookedUp := s.cloneBuildID(&Signal{
		InstallID:   "inl-1",
		ComponentID: "cmp-1",
		DeployID:    "dpl-requested",
		BuildID:     "bld-pinned",
	})
	require.Equal(s.T(), "bld-pinned", got)
	require.Empty(s.T(), lookedUp)
}
