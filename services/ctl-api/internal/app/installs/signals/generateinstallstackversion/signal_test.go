package generateinstallstackversion

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

type SignalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestSignalTestSuite(t *testing.T) {
	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) TestExecute() {
	s.T().Skip("not yet implemented")
}

// TestSetStackVersionCompositeRenderError_Enabled verifies that when composite
// errors are enabled the helper fires the recording activity with the right fields.
func (s *SignalTestSuite) TestSetStackVersionCompositeRenderError_Enabled() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	var capturedReq activities.SetInstallStackVersionCompositeErrorRequest
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args converter.EncodedValues) {
		if info.ActivityType.Name == "SetInstallStackVersionCompositeError" {
			require.NoError(s.T(), args.Get(&capturedReq))
		}
	})

	env.OnActivity((*activities.Activities).SetInstallStackVersionCompositeError, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		setStackVersionCompositeRenderError(ctx, "sv-1", "aws", "bad template var: {{.Missing}}", true)
		return nil
	})

	require.NoError(s.T(), env.GetWorkflowError())
	require.Equal(s.T(), "sv-1", capturedReq.StackVersionID)
	require.Equal(s.T(), "aws", capturedReq.Platform)
	require.Equal(s.T(), "bad template var: {{.Missing}}", capturedReq.Detail)
	env.AssertExpectations(s.T())
}

// TestSetStackVersionCompositeRenderError_Disabled verifies that when composite
// errors are disabled the helper skips the recording activity entirely.
func (s *SignalTestSuite) TestSetStackVersionCompositeRenderError_Disabled() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		setStackVersionCompositeRenderError(ctx, "sv-1", "aws", "some error", false)
		return nil
	})

	require.NoError(s.T(), env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

// TestSetStackVersionCompositeRenderError_ActivityError verifies that when the
// recording activity returns an error the helper swallows it (logged as warn)
// so it doesn't shadow the original render failure.
func (s *SignalTestSuite) TestSetStackVersionCompositeRenderError_ActivityError() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	env.OnActivity((*activities.Activities).SetInstallStackVersionCompositeError, mock.Anything, mock.Anything, mock.Anything).
		Return(temporal.NewNonRetryableApplicationError("db error", "TEST", nil)).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		setStackVersionCompositeRenderError(ctx, "sv-1", "gcp", "render error", true)
		return nil
	})

	require.NoError(s.T(), env.GetWorkflowError(), "helper must swallow recording errors")
	env.AssertExpectations(s.T())
}

func TestIsLegacyGCPInitScript(t *testing.T) {
	tests := map[string]struct {
		url  string
		want bool
	}{
		"default script": {url: DefaultGCPRunnerInitScript, want: true},
		"suffix match":   {url: "https://cdn.example.com/scripts/gcp/init.sh", want: true},
		"v2 script":      {url: "https://cdn.example.com/scripts/gcp/init-mng-v2.sh", want: false},
		"unrelated url":  {url: "https://example.com/setup.sh", want: false},
		"empty":          {url: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isLegacyGCPInitScript(tt.url)
			require.Equal(t, tt.want, got, "isLegacyGCPInitScript(%q)", tt.url)
		})
	}
}
