package syncer

import (
	"context"
	"errors"

	"github.com/nuonco/nuon/pkg/config"
	pkgsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/installs"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

const (
	e2eRunnerAPI      = "https://runner.example.com"
	e2ePublicAPI      = "https://api.example.com"
	e2eInitScript     = "https://raw.githubusercontent.com/acme/runner/main/init.sh"
	e2eAllowedVPC     = "https://acme-templates.s3.us-west-2.amazonaws.com/vpc/stack.yaml"
	e2eAllowedRunner  = "https://acme-templates.s3.us-west-2.amazonaws.com/runner/stack.yaml"
	e2eOffListVPC     = "https://mallory-templates.s3.us-west-2.amazonaws.com/vpc/stack.yaml"
	e2eOffListRunner  = "https://mallory-templates.s3.us-west-2.amazonaws.com/runner/stack.yaml"
	e2eOffListNested  = "https://mallory-templates.s3.us-west-2.amazonaws.com/nested.yaml"
	e2eLookalikeInit  = "https://raw.githubusercontent.com.evil.test/acme/runner/main/init.sh"
	e2eOtherOwnerInit = "https://raw.githubusercontent.com/mallory/runner/main/init.sh"
)

func (s *SyncFieldsTestSuite) e2ePolicy(mode urlpolicy.Mode) *urlpolicy.Policy {
	policy, err := urlpolicy.New(urlpolicy.Config{
		Mode: mode,
		Categories: map[urlpolicy.Category]urlpolicy.CategoryConfig{
			urlpolicy.CategoryRunnerAPI:       {Entries: []string{e2eRunnerAPI}},
			urlpolicy.CategoryPublicAPI:       {Entries: []string{e2ePublicAPI}},
			urlpolicy.CategoryRunnerBootstrap: {Entries: []string{"https://raw.githubusercontent.com/acme/runner/"}},
			urlpolicy.CategoryStackTemplates:  {Entries: []string{"https://acme-templates.s3.us-west-2.amazonaws.com/"}},
		},
	})
	s.Require().NoError(err)
	return policy
}

func e2eAllowedConfig() *config.AppConfig {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Runner.RunnerAPIURL = e2eRunnerAPI
	cfg.Runner.PublicAPIURL = e2ePublicAPI
	cfg.Runner.InitScriptURL = e2eInitScript + "#default"
	cfg.Runner.EnvVarMap = map[string]string{"RUNNER_AUTH_METHOD": "aws-sts", "LOG_LEVEL": "debug"}
	cfg.Stack = &config.StackConfig{
		Type:                    string(app.StackTypeAWS),
		Name:                    "stack",
		Description:             "stack",
		VPCNestedTemplateURL:    e2eAllowedVPC,
		RunnerNestedTemplateURL: e2eAllowedRunner,
	}
	return cfg
}

func (s *SyncFieldsTestSuite) installConfigCount(installID string) int64 {
	var count int64
	s.Require().NoError(s.deps.DB.Model(&app.InstallConfig{}).
		Where(&app.InstallConfig{InstallID: installID}).Count(&count).Error)
	return count
}

func (s *SyncFieldsTestSuite) installCount(appID, name string) int64 {
	var count int64
	s.Require().NoError(s.deps.DB.Model(&app.Install{}).
		Where(&app.Install{AppID: appID, Name: name}).Count(&count).Error)
	return count
}

func (s *SyncFieldsTestSuite) requirePolicyUserError(err error, field, setting string) {
	s.Require().Error(err)
	var userErr stderr.ErrUser
	s.Require().True(errors.As(err, &userErr), "a policy violation is a user error, got %T: %v", err, err)
	s.Contains(userErr.Description, field)
	s.Contains(userErr.Description, setting)
}

func (s *SyncFieldsTestSuite) TestURLPolicyEndToEnd() {
	original := s.deps.Cfg.URLPolicy
	defer func() { s.deps.Cfg.URLPolicy = original }()

	enforce := s.e2ePolicy(urlpolicy.ModeEnforce)
	s.deps.Cfg.URLPolicy = enforce

	ctx, testApp, _ := s.syncEmpty()

	s.Run("enforce: an app config inside every list syncs", func() {
		appCfg := s.syncInto(ctx, testApp.ID, e2eAllowedConfig(), WithURLPolicy(enforce))

		var runnerCfg app.AppRunnerConfig
		s.Require().NoError(s.deps.DB.Where(&app.AppRunnerConfig{AppConfigID: appCfg.ID}).First(&runnerCfg).Error)
		s.Equal(e2eRunnerAPI, runnerCfg.RunnerAPIURL)
		s.Equal(e2ePublicAPI, runnerCfg.PublicAPIURL)
		s.Equal(e2eInitScript+"#default", runnerCfg.InitScriptURL)

		var stackCfg app.AppStackConfig
		s.Require().NoError(s.deps.DB.Where(&app.AppStackConfig{AppConfigID: appCfg.ID}).First(&stackCfg).Error)
		s.Equal(e2eAllowedVPC, stackCfg.VPCNestedTemplateURL)
		s.Equal(e2eAllowedRunner, stackCfg.RunnerNestedTemplateURL)
	})

	s.Run("enforce: an app config outside a list is rejected and rolled back", func() {
		tests := map[string]struct {
			mutate   func(*config.AppConfig)
			resource string
			field    string
			reason   string
			setting  string
		}{
			"runner api on another host": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.RunnerAPIURL = "https://evil.test" },
				resource: "app-runner", field: "runner.runner_api_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_runner_api",
			},
			"runner api over http": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.RunnerAPIURL = "http://runner.example.com" },
				resource: "app-runner", field: "runner.runner_api_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_runner_api",
			},
			"public api on another host": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.PublicAPIURL = "https://evil.test" },
				resource: "app-runner", field: "runner.public_api_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_public_api",
			},
			"init script on a lookalike host": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.InitScriptURL = e2eLookalikeInit },
				resource: "app-runner", field: "runner.init_script_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_runner_bootstrap",
			},
			"init script from another owner on a shared host": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.InitScriptURL = e2eOtherOwnerInit },
				resource: "app-runner", field: "runner.init_script_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_runner_bootstrap",
			},
			"init script on another port": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.InitScriptURL = "https://raw.githubusercontent.com:8443/acme/runner/main/init.sh"
				},
				resource: "app-runner", field: "runner.init_script_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_runner_bootstrap",
			},
			"init script with dot dot segments": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.InitScriptURL = "https://raw.githubusercontent.com/acme/runner/../../mallory/init.sh"
				},
				resource: "app-runner", field: "runner.init_script_url",
				reason: `".."`, setting: "allowed_hosts_runner_bootstrap",
			},
			"init script with an encoded slash": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.InitScriptURL = "https://raw.githubusercontent.com/acme/runner/..%2f..%2fmallory/init.sh"
				},
				resource: "app-runner", field: "runner.init_script_url",
				reason: "encoded slash", setting: "allowed_hosts_runner_bootstrap",
			},
			"init script with a shell metacharacter": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.InitScriptURL = e2eInitScript + ";id" },
				resource: "app-runner", field: "runner.init_script_url",
				reason: "must not contain the character", setting: "allowed_hosts_runner_bootstrap",
			},
			"init script with command substitution": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.InitScriptURL = e2eInitScript + "?a=$(id)" },
				resource: "app-runner", field: "runner.init_script_url",
				reason: "must not contain the character", setting: "allowed_hosts_runner_bootstrap",
			},
			"init script with a trailing newline": {
				mutate:   func(cfg *config.AppConfig) { cfg.Runner.InitScriptURL = e2eInitScript + "\n" },
				resource: "app-runner", field: "runner.init_script_url",
				reason: "whitespace", setting: "allowed_hosts_runner_bootstrap",
			},
			"phone home script with credentials": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.PhoneHomeScriptURL = "https://u:p@raw.githubusercontent.com/acme/runner/main/phone.py"
				},
				resource: "app-runner", field: "runner.phone_home_script_url",
				reason: "credentials", setting: "allowed_hosts_runner_bootstrap",
			},
			"reserved proxy env var": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.EnvVarMap = map[string]string{"HTTPS_PROXY": "http://evil.test:3128"}
				},
				resource: "app-runner", field: "runner.env_vars.HTTPS_PROXY",
				reason: "reserved", setting: "allowed_hosts_runner_api",
			},
			"reserved runner env var": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.EnvVarMap = map[string]string{"RUNNER_API_URL": "https://evil.test"}
				},
				resource: "app-runner", field: "runner.env_vars.RUNNER_API_URL",
				reason: "reserved", setting: "allowed_hosts_runner_api",
			},
			"reserved certificate env var": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Runner.EnvVarMap = map[string]string{"SSL_CERT_FILE": "/tmp/ca.pem"}
				},
				resource: "app-runner", field: "runner.env_vars.SSL_CERT_FILE",
				reason: "reserved", setting: "allowed_hosts_runner_api",
			},
			"vpc template in another bucket": {
				mutate:   func(cfg *config.AppConfig) { cfg.Stack.VPCNestedTemplateURL = e2eOffListVPC },
				resource: "app-cloudformation-stack", field: "stack.vpc_nested_template_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_stack_templates",
			},
			"runner template in another bucket": {
				mutate:   func(cfg *config.AppConfig) { cfg.Stack.RunnerNestedTemplateURL = e2eOffListRunner },
				resource: "app-cloudformation-stack", field: "stack.runner_nested_template_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_stack_templates",
			},
			"custom nested stack in another bucket": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Stack.CustomNestedStacks = []config.CustomNestedStack{{Name: "extra", TemplateURL: e2eOffListNested, Contents: "{}"}}
				},
				resource: "app-cloudformation-stack", field: "stack.custom_nested_stacks[0].template_url",
				reason: "does not match any allowed entry", setting: "allowed_hosts_stack_templates",
			},
			"custom nested stack with a scheme relative url": {
				mutate: func(cfg *config.AppConfig) {
					cfg.Stack.CustomNestedStacks = []config.CustomNestedStack{{Name: "extra", TemplateURL: "//mallory.test/nested.yaml", Contents: "{}"}}
				},
				resource: "app-cloudformation-stack", field: "stack.custom_nested_stacks[0].template_url",
				reason: "absolute https URL", setting: "allowed_hosts_stack_templates",
			},
		}
		for name, tt := range tests {
			s.Run(name, func() {
				cfg := e2eAllowedConfig()
				tt.mutate(cfg)

				appCfg, err := s.trySync(ctx, testApp.ID, cfg, WithURLPolicy(enforce))
				s.Require().Error(err)

				var syncErr pkgsync.SyncErr
				s.Require().True(errors.As(err, &syncErr), "a policy violation is a user-facing sync error, got %T: %v", err, err)
				s.Equal(tt.resource, syncErr.Resource)
				s.Contains(syncErr.Description, tt.field)
				s.Contains(syncErr.Description, tt.reason)
				s.Contains(syncErr.Description, tt.setting)
				s.Zero(s.runnerConfigCount(appCfg.ID), "the sync transaction must roll back")
			})
		}
	})

	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	s.Run("enforce: install stack overrides inside the list are stored", func() {
		created, err := s.deps.InstallHelpers.CreateInstallConfig(ctx, install.ID, &installhelpers.CreateInstallConfigParams{
			ApprovalOption:          app.InstallApprovalOptionPrompt,
			VPCNestedTemplateURL:    generics.ToPtr(e2eAllowedVPC),
			RunnerNestedTemplateURL: generics.ToPtr(e2eAllowedRunner),
		})
		s.Require().NoError(err)
		s.Require().NotNil(created.RunnerNestedTemplateURL)
		s.Equal(e2eAllowedRunner, *created.RunnerNestedTemplateURL)
		s.NoError(s.deps.InstallHelpers.CheckStackOverridesURLPolicy(ctx, install.ID, generics.ToPtr(e2eAllowedVPC), nil, nil))
	})

	s.Run("enforce: install stack overrides outside the list are rejected", func() {
		before := s.installConfigCount(install.ID)

		tests := map[string]struct {
			params installhelpers.CreateInstallConfigParams
			field  string
		}{
			"vpc template": {
				params: installhelpers.CreateInstallConfigParams{VPCNestedTemplateURL: generics.ToPtr(e2eOffListVPC)},
				field:  "stack_overrides.vpc_nested_template_url",
			},
			"runner template": {
				params: installhelpers.CreateInstallConfigParams{RunnerNestedTemplateURL: generics.ToPtr(e2eOffListRunner)},
				field:  "stack_overrides.runner_nested_template_url",
			},
			"custom nested stack": {
				params: installhelpers.CreateInstallConfigParams{
					CustomNestedStacks: []config.CustomNestedStack{{Name: "extra", TemplateURL: e2eOffListNested}},
				},
				field: "stack_overrides.custom_nested_stacks[0].template_url",
			},
		}
		for name, tt := range tests {
			s.Run(name, func() {
				params := tt.params
				params.ApprovalOption = app.InstallApprovalOptionPrompt

				_, err := s.deps.InstallHelpers.CreateInstallConfig(ctx, install.ID, &params)
				s.requirePolicyUserError(err, tt.field, "allowed_hosts_stack_templates")

				err = s.deps.InstallHelpers.CheckStackOverridesURLPolicy(ctx, install.ID, params.VPCNestedTemplateURL, params.RunnerNestedTemplateURL, params.CustomNestedStacks)
				s.requirePolicyUserError(err, tt.field, "allowed_hosts_stack_templates")
			})
		}
		s.Equal(before, s.installConfigCount(install.ID), "a rejected override must not create an install config")
	})

	s.Run("enforce: install sync rejects an off-list override on an existing install before writing", func() {
		before := s.installConfigCount(install.ID)

		_, err := installs.SyncInstall(ctx, s.deps.DB, s.deps.InstallHelpers, testApp.ID, &config.Install{
			Name:           install.Name,
			StackOverrides: &config.InstallStackOverrides{RunnerNestedTemplateURL: e2eOffListRunner},
		})
		s.requirePolicyUserError(err, "stack_overrides.runner_nested_template_url", "allowed_hosts_stack_templates")
		s.Equal(before, s.installConfigCount(install.ID))

		var latest app.InstallConfig
		s.Require().NoError(s.deps.DB.Where(&app.InstallConfig{InstallID: install.ID}).Order("created_at DESC").First(&latest).Error)
		s.Require().NotNil(latest.RunnerNestedTemplateURL)
		s.Equal(e2eAllowedRunner, *latest.RunnerNestedTemplateURL)
	})

	s.Run("enforce: install sync rejects an off-list override on a new install before creating it", func() {
		_, err := installs.SyncInstall(ctx, s.deps.DB, s.deps.InstallHelpers, testApp.ID, &config.Install{
			Name:           "policy-new-install",
			AWSAccount:     &config.AWSAccount{Region: "us-west-2"},
			StackOverrides: &config.InstallStackOverrides{VPCNestedTemplateURL: e2eOffListVPC},
		})
		s.requirePolicyUserError(err, "stack_overrides.vpc_nested_template_url", "allowed_hosts_stack_templates")
		s.Zero(s.installCount(testApp.ID, "policy-new-install"), "a rejected override must not leave an install behind")
	})

	s.Run("warn: violations are reported and nothing is rejected", func() {
		warn := s.e2ePolicy(urlpolicy.ModeWarn)
		var reported urlpolicy.Violations
		warn.OnWarn(func(v urlpolicy.Violations) { reported = append(reported, v...) })
		s.deps.Cfg.URLPolicy = warn
		defer func() { s.deps.Cfg.URLPolicy = enforce }()

		cfg := e2eAllowedConfig()
		cfg.Runner.RunnerAPIURL = "https://evil.test"
		cfg.Runner.InitScriptURL = e2eLookalikeInit
		cfg.Runner.EnvVarMap = map[string]string{"HTTPS_PROXY": "http://evil.test:3128"}
		cfg.Stack.VPCNestedTemplateURL = e2eOffListVPC
		appCfg := s.syncInto(ctx, testApp.ID, cfg, WithURLPolicy(warn))
		s.EqualValues(1, s.runnerConfigCount(appCfg.ID))

		warnInstall := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)
		_, err := s.deps.InstallHelpers.CreateInstallConfig(ctx, warnInstall.ID, &installhelpers.CreateInstallConfigParams{
			ApprovalOption:          app.InstallApprovalOptionPrompt,
			RunnerNestedTemplateURL: generics.ToPtr(e2eOffListRunner),
		})
		s.Require().NoError(err)

		fields := make([]string, 0, len(reported))
		for _, violation := range reported {
			fields = append(fields, violation.Field)
		}
		s.ElementsMatch([]string{
			"runner.runner_api_url",
			"runner.init_script_url",
			"runner.env_vars.HTTPS_PROXY",
			"stack.vpc_nested_template_url",
			"stack_overrides.runner_nested_template_url",
		}, fields)
	})

	s.Run("unrestricted: the same values are accepted without a policy", func() {
		s.deps.Cfg.URLPolicy = nil
		defer func() { s.deps.Cfg.URLPolicy = enforce }()

		cfg := e2eAllowedConfig()
		cfg.Runner.RunnerAPIURL = "https://evil.test"
		cfg.Runner.InitScriptURL = e2eLookalikeInit + ";id"
		cfg.Runner.EnvVarMap = map[string]string{"HTTPS_PROXY": "http://evil.test:3128"}
		cfg.Stack.VPCNestedTemplateURL = e2eOffListVPC
		appCfg := s.syncInto(ctx, testApp.ID, cfg)
		s.EqualValues(1, s.runnerConfigCount(appCfg.ID))

		openInstall := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)
		_, err := s.deps.InstallHelpers.CreateInstallConfig(ctx, openInstall.ID, &installhelpers.CreateInstallConfigParams{
			ApprovalOption:       app.InstallApprovalOptionPrompt,
			VPCNestedTemplateURL: generics.ToPtr(e2eOffListVPC),
		})
		s.Require().NoError(err)
	})

	s.Run("enforce again: values stored while unrestricted are refused on the next write", func() {
		cfg := e2eAllowedConfig()
		cfg.Runner.RunnerAPIURL = "https://evil.test"
		_, err := s.trySync(context.WithoutCancel(ctx), testApp.ID, cfg, WithURLPolicy(enforce))
		s.Require().Error(err)

		lateInstall := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)
		_, err = s.deps.InstallHelpers.CreateInstallConfig(ctx, lateInstall.ID, &installhelpers.CreateInstallConfigParams{
			ApprovalOption:       app.InstallApprovalOptionPrompt,
			VPCNestedTemplateURL: generics.ToPtr(e2eOffListVPC),
		})
		s.Zero(s.installConfigCount(lateInstall.ID))
		s.requirePolicyUserError(err, "stack_overrides.vpc_nested_template_url", "allowed_hosts_stack_templates")
	})
}
