package updateinstallstackoutputs

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "update-install-stack-outputs"

type Signal struct {
	InstallStackID          string
	InstallStackVersionID   string
	SkipInputUpdateWorkflow bool
}

var (
	_ signal.Signal              = &Signal{}
	_ signal.SignalWithAutoRetry = (*Signal)(nil)
)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallStackID == "" {
		return fmt.Errorf("install stack id is required")
	}

	// Validate install stack exists
	_, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
	if err != nil {
		return fmt.Errorf("unable to get install for stack: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	var err error
	var install *app.Install
	var version *app.InstallStackVersion

	if s.InstallStackVersionID != "" {
		version, err = activities.AwaitGetInstallStackVersionByID(ctx, activities.GetInstallStackVersionByIDRequest{
			VersionID: s.InstallStackVersionID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to get install stack version")
		}
		install, err = activities.AwaitGetInstallForStackByStackID(ctx, version.InstallStackID)
		if err != nil {
			return errors.Wrap(err, "unable to get install")
		}
	} else {
		install, err = activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
		if err != nil {
			return errors.Wrap(err, "unable to get install")
		}
		version, err = activities.AwaitGetInstallStackVersionByInstallID(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install version")
		}
	}

	run, err := activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get run outputs")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config by id")
	}

	installStackOutputs := app.InstallStackOutputs{}

	var stackOutputs app.StackOutput
	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		installStackOutputs.AWSStackOutputs = &app.AWSStackOutputs{}
		stackOutputs = installStackOutputs.AWSStackOutputs
	case app.AppRunnerTypeAzure:
		installStackOutputs.AzureStackOutputs = &app.AzureStackOutputs{}
		stackOutputs = installStackOutputs.AzureStackOutputs
	case app.AppRunnerTypeGCP:
		installStackOutputs.GCPStackOutputs = &app.GCPStackOutputs{}
		stackOutputs = installStackOutputs.GCPStackOutputs
	default:
		return nil
	}

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			pkggenerics.StringToMapDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           stackOutputs,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrapf(err, "unable to create %s decoder", appCfg.RunnerConfig.Type)
	}
	if err := decoder.Decode(run.Data); err != nil {
		return errors.Wrapf(err, "unable to parse %s install outputs", appCfg.RunnerConfig.Type)
	}

	// update outputs if needed
	if err := activities.AwaitUpdateInstallStackOutputs(ctx, activities.UpdateInstallStackOutputs{
		InstallStackID:           version.InstallStackID,
		InstallStackVersionRunID: run.ID,
		Data:                     generics.ToStringMap(run.Data),
	}); err != nil {
		return errors.Wrap(err, "unable to update install stack outputs")
	}

	// update install roles provisioned state from stack outputs
	if err := activities.AwaitUpdateInstallRolesFromStackOutputs(ctx, activities.UpdateInstallRolesFromStackOutputs{
		InstallID: install.ID,
	}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Warn("unable to update install roles from stack outputs", zap.Error(err))
	}

	// update the runner settings group
	runnerIAMRoleARN := ""
	if installStackOutputs.AWSStackOutputs != nil {
		runnerIAMRoleARN = installStackOutputs.AWSStackOutputs.RunnerIAMRoleARN
	}

	if err := activities.AwaitUpdateRunnerGroupSettings(ctx, &activities.UpdateRunnerGroupSettings{
		RunnerID:           install.RunnerID,
		LocalAWSIAMRoleARN: runnerIAMRoleARN,
	}); err != nil {
		return errors.Wrap(err, "unable to update runner group settings")
	}

	// update gcp account from stack outputs
	if installStackOutputs.GCPStackOutputs != nil && installStackOutputs.GCPStackOutputs.Region != "" {
		if err := activities.AwaitUpdateGCPAccountRegion(ctx, &activities.UpdateGCPAccountRegion{
			InstallID: install.ID,
			Region:    installStackOutputs.GCPStackOutputs.Region,
			ProjectID: installStackOutputs.GCPStackOutputs.ProjectID,
		}); err != nil {
			return errors.Wrap(err, "unable to update gcp account from stack outputs")
		}
	}

	// NOTE(jm): this is probably not the _best_ place to do this validation, but for now it works
	// make sure the region matches the outputs
	err = validateRegion(*install, installStackOutputs)
	if err != nil {
		return errors.Wrap(err, "unable to validate region")
	}

	installInputValues, err := stackOutputs.InstallInputValues()
	if err != nil {
		return errors.Wrap(err, "unable to fetch install input values from stack outputs")
	}
	if len(installInputValues) > 0 {
		if err := activities.AwaitUpdateInstallInputsFromStack(ctx, &activities.UpdateInstallInputsFromStackRequest{
			InstallID:               install.ID,
			InputConfigID:           appCfg.InputConfig.ID,
			InputValues:             installInputValues,
			InstallStackVersionID:   version.ID,
			SkipInputUpdateWorkflow: s.SkipInputUpdateWorkflow,
		}); err != nil {
			return errors.Wrap(err, "unable to update install inputs from stack outputs")
		}
	}

	if _, err := state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   run.ID,
		TriggeredByType: "install_stack_outputs",
	}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Warn("unable to generate state", zap.Error(err))
	}

	return nil
}

func validateRegion(install app.Install, outputs app.InstallStackOutputs) error {
	// Region is recorded on the install only when the user committed to one
	// at install creation. AWS/Azure both allow creating an install with no
	// region, in which case the user picks one at apply time and the stack
	// outputs are the source of truth. Only enforce a match when the install
	// already carries a region.
	switch {
	case install.AWSAccount != nil:
		if install.AWSAccount.Region != "" && install.AWSAccount.Region != outputs.AWSStackOutputs.Region {
			return errors.New("install stack was run for a different region than the install was configured for")
		}
	case install.AzureAccount != nil:
		if install.AzureAccount.Location != "" && install.AzureAccount.Location != outputs.AzureStackOutputs.ResourceGroupLocation {
			return errors.New("install stack was run for a different region than the install was configured for")
		}
	}

	return nil
}
