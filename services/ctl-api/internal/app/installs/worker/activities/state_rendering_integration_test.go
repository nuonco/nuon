package activities_test

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/pkg/generics"
	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/rolechange"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/stackrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (s *InstallStateIntegrationSuite) TestOutOfBandStackRunRendersRoleNameFromUpdatedInputs() {
	roleCfg := &app.AppAWSIAMRoleConfig{
		AppConfigID: s.install.AppConfigID,
		Type:        app.AWSIAMRoleTypeRunnerProvision,
		Name:        "role-{{.nuon.inputs.inputs.region}}",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(roleCfg).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.InstallRoles{
		InstallID: s.install.ID, AppRoleConfigID: roleCfg.ID, Enabled: true,
	}).Error)
	version := s.deps.Seeder.CreateInstallStackVersion(s.ctx, s.T(), s.install.ID, s.install.InstallStack.ID, s.install.AppConfigID)
	run := &app.InstallStackVersionRun{
		InstallStackVersionID: version.ID,
		Data: pgtype.Hstore{
			"region":                 generics.ToPtr(s.install.AWSAccount.Region),
			"provision_iam_role_arn": generics.ToPtr("arn:aws:iam::123456789012:role/provision"),
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(run).Error)

	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{BackgroundActivityContext: s.ctx, DeadlockDetectionTimeout: time.Minute})
	env.SetTestTimeout(2 * time.Minute)
	env.SetDataConverter(converter.NewCompositeDataConverter(
		signaldb.NewPayloadConverter(),
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))
	env.RegisterActivity(s.activities)
	env.OnActivity((*installactivities.Activities).ReconcileRunnerEnabled, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity((*installactivities.Activities).UpdateRunnerGroupSettings, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	var roleChanges []*rolechange.Signal
	env.OnActivity(new(sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, req *sharedactivities.EnqueueSignalToOwnerRequest) (*sharedactivities.EnqueueSignalToOwnerResponse, error) {
			switch sig := req.Signal.(type) {
			case *statepartialgenerate.Signal:
				regenerated, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
				if err != nil {
					return nil, err
				}
				require.Equal(s.T(), "us-east-1", regenerated.Inputs.Inputs["region"])
				s.updateInput()
			case *rolechange.Signal:
				roleChanges = append(roleChanges, sig)
			}
			return &sharedactivities.EnqueueSignalToOwnerResponse{}, nil
		})
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(callback.SignalName(s.install.ID), callback.Result{Status: "ok"})
	}, 0)

	sig := &stackrun.Signal{InstallStackID: s.install.InstallStack.ID, InstallStackVersionID: version.ID, RunID: run.ID}
	env.ExecuteWorkflow(sig.Execute)

	require.True(s.T(), env.IsWorkflowCompleted())
	require.NoError(s.T(), env.GetWorkflowError())
	current, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "eu-west-1", current.Inputs.Inputs["region"])
	require.Len(s.T(), roleChanges, 1)
	require.Equal(s.T(), "enabled", roleChanges[0].ChangeType)
	require.Equal(s.T(), "role-eu-west-1", roleChanges[0].RoleName,
		"out-of-band role-change must be rendered with the install's current inputs")
}
