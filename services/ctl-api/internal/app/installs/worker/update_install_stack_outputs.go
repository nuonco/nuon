package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen workflow
// @execution-timeout 24h
// @task-timeout 30s
func (w *Workflows) UpdateInstallStackOutputs(ctx workflow.Context, sreq signals.RequestSignal) error {
	version, err := activities.AwaitGetInstallStackVersionByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install version")
	}

	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	run, err := activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get run outputs")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config by id")
	}

	if appCfg.RunnerConfig.Type != app.AppRunnerTypeAWS {
		return nil
	}

	// make sure outputs are valid
	var outputs app.AWSStackOutputs

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
		),
		WeaklyTypedInput: true,
		Result:           &outputs,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create decoder")
	}
	if err := decoder.Decode(run.Data); err != nil {
		return errors.Wrap(err, "unable to parse aws outputs")
	}

	if err := w.v.Struct(outputs); err != nil {
		return errors.Wrap(err, "invalid outputs")
	}

	// update outputs if needed
	if err := activities.AwaitUpdateInstallStackOutputs(ctx, activities.UpdateInstallStackOutputs{
		InstallStackID:           version.InstallStackID,
		InstallStackVersionRunID: version.ID,
		Data:                     generics.ToStringMap(run.Data),
	}); err != nil {
		return errors.Wrap(err, "unable to update install stack outputs")
	}

	return nil
}
