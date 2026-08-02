package build

import (
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RunnerInput struct {
	AppID       string
	AppConfigID string

	Type          string
	HelmDriver    string
	EnvVars       map[string]string
	InitScriptURL string
	InstanceType  string
	RunnerAPIURL  string
	PublicAPIURL  string

	PhoneHomeScriptURL string
}

func RunnerInputFromConfig(runner config.AppRunnerConfig, appID, appConfigID string) RunnerInput {
	return RunnerInput{
		AppID:         appID,
		AppConfigID:   appConfigID,
		Type:          runner.RunnerType,
		HelmDriver:    runner.HelmDriver,
		EnvVars:       runner.EnvVarMap,
		InitScriptURL: runner.InitScriptURL,
		InstanceType:  runner.InstanceType,
		RunnerAPIURL:  runner.RunnerAPIURL,
		PublicAPIURL:  runner.PublicAPIURL,

		PhoneHomeScriptURL: runner.PhoneHomeScriptURL,
	}
}

func RunnerConfig(in RunnerInput) *app.AppRunnerConfig {
	return &app.AppRunnerConfig{
		AppID:         in.AppID,
		AppConfigID:   in.AppConfigID,
		Type:          app.AppRunnerType(in.Type),
		HelmDriver:    app.AppRunnerConfigHelmDriverType(in.HelmDriver),
		EnvVars:       Hstore(in.EnvVars),
		InitScriptURL: in.InitScriptURL,
		InstanceType:  in.InstanceType,
		RunnerAPIURL:  in.RunnerAPIURL,
		PublicAPIURL:  in.PublicAPIURL,

		PhoneHomeScriptURL: in.PhoneHomeScriptURL,
	}
}
