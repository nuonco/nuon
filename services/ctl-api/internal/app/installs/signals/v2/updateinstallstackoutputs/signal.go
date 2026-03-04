package updateinstallstackoutputs

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/types/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "update-install-stack-outputs"

type Signal struct {
	InstallStackID string
}

var _ signal.Signal = &Signal{}

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
	install, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	version, err := activities.AwaitGetInstallStackVersionByInstallID(ctx, install.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install version")
	}

	run, err := activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get run outputs")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config by id")
	}

	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		break
	case app.AppRunnerTypeAzure:
		break
	default:
		return nil
	}

	// make sure outputs are valid
	outputs := app.InstallStackOutputs{
		AWSStackOutputs:   nil,
		AzureStackOutputs: nil,
	}

	v := validator.New()

	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		// parse into map[string]interface{}
		stackOutputs, err := stacks.DecodeAWSStackOutputData(run.Data)
		if err != nil {
			return errors.Wrap(err, "unable to decode run data")
		}

		// parse into AWSStackOutputs
		decoderConfig := &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				mapstructure.StringToTimeDurationHookFunc(),
			),
			WeaklyTypedInput: true,
			Result:           &outputs.AWSStackOutputs,
		}
		decoder, err := mapstructure.NewDecoder(decoderConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create decoder")
		}
		if err := decoder.Decode(stackOutputs); err != nil {
			return errors.Wrap(err, "unable to parse install outputs")
		}

		if err := v.Struct(outputs); err != nil {
			return errors.Wrap(err, "invalid outputs")
		}
	case app.AppRunnerTypeAzure:
		decoderConfig := &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				mapstructure.StringToTimeDurationHookFunc(),
			),
			WeaklyTypedInput: true,
			Result:           &outputs.AzureStackOutputs,
		}
		decoder, err := mapstructure.NewDecoder(decoderConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create decoder")
		}
		if err := decoder.Decode(run.Data); err != nil {
			return errors.Wrap(err, "unable to parse install outputs")
		}

		if err := v.Struct(outputs); err != nil {
			return errors.Wrap(err, "invalid outputs")
		}
	}

	// update outputs if needed
	if err := activities.AwaitUpdateInstallStackOutputs(ctx, activities.UpdateInstallStackOutputs{
		InstallStackID:           version.InstallStackID,
		InstallStackVersionRunID: run.ID,
		Data:                     generics.ToStringMap(run.Data),
	}); err != nil {
		return errors.Wrap(err, "unable to update install stack outputs")
	}

	// update the runner settings group
	runnerIAMRoleARN := ""
	if outputs.AWSStackOutputs != nil {
		runnerIAMRoleARN = outputs.AWSStackOutputs.RunnerIAMRoleARN
	}
	if err := activities.AwaitUpdateRunnerGroupSettings(ctx, &activities.UpdateRunnerGroupSettings{
		RunnerID:           install.RunnerID,
		LocalAWSIAMRoleARN: runnerIAMRoleARN,
	}); err != nil {
		return errors.Wrap(err, "unable to update runner group settings")
	}

	// NOTE(jm): this is probably not the _best_ place to do this validation, but for now it works
	// make sure the region matches the outputs
	err = validateRegion(*install, outputs)
	if err != nil {
		return errors.Wrap(err, "unable to validate region")
	}

	// TODO(sk): this can be aws or azure
	data, err := stacks.DecodeAWSStackOutputData(run.Data)
	if err != nil {
		return errors.Wrap(err, "unable to decode run data")
	}

	// extract app_inputs from stack outputs for install_stack sourced inputs
	inputValues := extractAppInputsFromStackOutputs(data)
	if len(inputValues) > 0 {
		if err := activities.AwaitUpdateInstallInputsFromStack(ctx, &activities.UpdateInstallInputsFromStackRequest{
			InstallID:             install.ID,
			InputConfigID:         appCfg.InputConfig.ID,
			InputValues:           inputValues,
			InstallStackVersionID: version.ID,
		}); err != nil {
			return errors.Wrap(err, "unable to update install inputs from stack outputs")
		}
	}

	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   run.ID,
		TriggeredByType: "install_stack_outputs",
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}

func validateRegion(install app.Install, outputs app.InstallStackOutputs) error {
	switch {
	case install.AWSAccount != nil:
		if install.AWSAccount.Region != outputs.AWSStackOutputs.Region {
			return errors.New("install stack was run for a different region than the install was configured for")
		}
	case install.AzureAccount != nil:
		if install.AzureAccount.Location != outputs.AzureStackOutputs.ResourceGroupLocation {
			return errors.New("install stack was run for a different region than the install was configured for")
		}
	}

	return nil
}

// extractAppInputsFromStackOutputs extracts the install_inputs nested object from stack outputs
func extractAppInputsFromStackOutputs(stackOutputs map[string]interface{}) map[string]string {
	inputValues := make(map[string]string)

	// Extract app_inputs from stack outputs
	appInputsRaw, ok := stackOutputs["install_inputs"]
	if !ok {
		return inputValues
	}

	// Convert app_inputs to map[string]interface{}
	appInputsMap, ok := appInputsRaw.(map[string]string)
	if !ok {
		return inputValues
	}

	return appInputsMap
}
