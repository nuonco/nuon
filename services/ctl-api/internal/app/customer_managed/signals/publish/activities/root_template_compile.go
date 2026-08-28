package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanaged "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

func (a *Activities) compileRootTemplate(ctx context.Context, cfg *app.AppConfig, syntheticInstallID, runnerImageTag string) ([]byte, string, error) {
	if syntheticInstallID == "" {
		return nil, "", fmt.Errorf("synthetic install ID is required")
	}

	cfg, err := customermanaged.RenderConfigForStackCompile(ctx, a.db, cfg, syntheticInstallID)
	if err != nil {
		return nil, "", fmt.Errorf("render app config for stack compile: %w", err)
	}

	phoneHomeScript, err := cloudformation.FetchPhoneHomeScript(ctx, cfg.RunnerConfig.PhoneHomeScriptURL, a.cfg.PhoneHomeScriptURL)
	if err != nil {
		return nil, "", err
	}

	instanceType := cfg.RunnerConfig.InstanceType
	if instanceType == "" {
		instanceType = app.DefaultInstanceTypeForPlatform(cfg.RunnerConfig.CloudPlatform)
	}
	initScriptURL := cfg.RunnerConfig.InitScriptURL
	if initScriptURL == "" {
		initScriptURL = defaultAWSRunnerInitScriptURL
	}

	install := &app.Install{ID: syntheticInstallID, OrgID: cfg.OrgID, AppID: cfg.AppID, AppConfigID: cfg.ID}
	runner := &app.Runner{ID: syntheticInstallID + "-runner", OrgID: cfg.OrgID}
	settings := &app.RunnerGroupSettings{AWSInstanceType: instanceType, RunnerAPIURL: cfg.RunnerConfig.RunnerAPIURL}
	publicAPIURL := cfg.RunnerConfig.PublicAPIURL
	if publicAPIURL == "" {
		publicAPIURL = a.cfg.PublicAPIURL
	}
	version := &app.InstallStackVersion{
		InstallID:    syntheticInstallID,
		AppConfigID:  cfg.ID,
		PhoneHomeID:  "compiled",
		PhoneHomeURL: fmt.Sprintf("%s/v1/installs/%s/phone-home/compiled", publicAPIURL, syntheticInstallID),
		StackName:    cfg.StackConfig.Name,
	}
	input := &stacks.TemplateInput{
		Install:                      install,
		CloudFormationStackVersion:   version,
		InstallState:                 &state.State{},
		AppCfg:                       cfg,
		Runner:                       runner,
		Settings:                     settings,
		APIToken:                     "compiled",
		RunnerEnvVars:                stacks.FormatRunnerEnvVars(&cfg.RunnerConfig, runnerImageTag),
		RunnerInitScriptURL:          initScriptURL,
		PhonehomeScript:              string(phoneHomeScript),
		VPCNestedStackTemplateURL:    cfg.StackConfig.VPCNestedTemplateURL,
		RunnerNestedStackTemplateURL: cfg.StackConfig.RunnerNestedTemplateURL,
	}

	template, checksum, err := cloudformation.NewTemplates(cloudformation.Params{Cfg: a.cfg}).Template(input)
	if err != nil {
		return nil, "", fmt.Errorf("compile root stack template: %w", err)
	}
	contents, err := template.JSON()
	if err != nil {
		return nil, "", fmt.Errorf("marshal compiled root stack template: %w", err)
	}
	contents, err = prepareRootTemplateForBundle(contents)
	if err != nil {
		return nil, "", fmt.Errorf("prepare compiled root stack template: %w", err)
	}
	return contents, "compiled:cloudformation:" + checksum, nil
}
