package build

import (
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
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

	ImageAutoUpdate bool
	ImageID         string
	PublicAPIURL    string

	PhoneHomeScriptURL string

	URLPolicy *urlpolicy.Policy
}

func RunnerInputFromConfig(runner config.AppRunnerConfig, appID, appConfigID string, policy *urlpolicy.Policy) RunnerInput {
	return RunnerInput{
		URLPolicy:     policy,
		AppID:         appID,
		AppConfigID:   appConfigID,
		Type:          runner.RunnerType,
		HelmDriver:    runner.HelmDriver,
		EnvVars:       runner.EnvVarMap,
		InitScriptURL: runner.InitScriptURL,
		InstanceType:  runner.InstanceType,
		RunnerAPIURL:  runner.RunnerAPIURL,
		PublicAPIURL:  runner.PublicAPIURL,

		ImageAutoUpdate: runner.ImageAutoUpdate,
		ImageID:         runner.ImageID,

		PhoneHomeScriptURL: runner.PhoneHomeScriptURL,
	}
}

func RunnerConfig(in RunnerInput) (*app.AppRunnerConfig, error) {
	if err := validation.CheckURLPolicy(in.URLPolicy, validate.RunnerURLPolicyViolations(in.URLPolicy, validate.RunnerURLs{
		RunnerAPIURL:       in.RunnerAPIURL,
		PublicAPIURL:       in.PublicAPIURL,
		InitScriptURL:      in.InitScriptURL,
		PhoneHomeScriptURL: in.PhoneHomeScriptURL,
		EnvVars:            in.EnvVars,
	})); err != nil {
		return nil, err
	}

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

		ImageAutoUpdate: in.ImageAutoUpdate,
		ImageID:         in.ImageID,

		PhoneHomeScriptURL: in.PhoneHomeScriptURL,
	}, nil
}
