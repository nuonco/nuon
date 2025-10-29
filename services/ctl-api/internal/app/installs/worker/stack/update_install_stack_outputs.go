package stack

import (
	"go.temporal.io/sdk/workflow"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/stacks"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

// @temporal-gen workflow
// @execution-timeout 24h
// @task-timeout 30s
func (w *Workflows) UpdateInstallStackOutputs(ctx workflow.Context, sreq signals.RequestSignal) error {
	id := sreq.ID
	if sreq.InstallStackID != "" {
		id = sreq.InstallStackID
	}

	install, err := activities.AwaitGetInstallForStackByStackID(ctx, id)
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

		if err := w.v.Struct(outputs); err != nil {
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

		if err := w.v.Struct(outputs); err != nil {
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
		TriggeredByType: plugins.TableName(w.db, outputs),
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

// extractAppInputsFromStackOutputs extracts the app_inputs nested object from stack outputs
func extractAppInputsFromStackOutputs(stackOutputs map[string]interface{}) map[string]string {
	inputValues := make(map[string]string)

	// Extract app_inputs from stack outputs
	appInputsRaw, ok := stackOutputs["app_inputs"]
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
