package syncer

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	pkgsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

func (s *SyncFieldsTestSuite) urlPolicy(mode urlpolicy.Mode) *urlpolicy.Policy {
	policy, err := urlpolicy.New(urlpolicy.Config{
		Mode: mode,
		Categories: map[urlpolicy.Category]urlpolicy.CategoryConfig{
			urlpolicy.CategoryRunnerAPI:       {Entries: []string{"https://runner.example.com"}},
			urlpolicy.CategoryRunnerBootstrap: {Entries: []string{"https://raw.githubusercontent.com/acme/runner/"}},
			urlpolicy.CategoryStackTemplates:  {Entries: []string{"https://acme-templates.s3.us-west-2.amazonaws.com/"}},
		},
	})
	s.Require().NoError(err)
	return policy
}

func (s *SyncFieldsTestSuite) trySync(ctx context.Context, appID string, cfg *config.AppConfig, opts ...Option) (*app.AppConfig, error) {
	appCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), appID)
	err := s.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return NewDBSyncer(
			tx,
			s.deps.AppsHelpers,
			s.deps.ComponentHelpers,
			s.deps.ActionsHelpers,
			s.deps.RunbooksHelpers,
			s.deps.InstallHelpers,
			s.deps.VCSHelpers,
			s.deps.TFClient,
			appID,
			cfg,
			appCfg.ID,
			opts...,
		).Sync(ctx)
	})
	return appCfg, err
}

func (s *SyncFieldsTestSuite) runnerConfigCount(appConfigID string) int64 {
	var count int64
	s.Require().NoError(s.deps.DB.Model(&app.AppRunnerConfig{}).
		Where(&app.AppRunnerConfig{AppConfigID: appConfigID}).Count(&count).Error)
	return count
}

func (s *SyncFieldsTestSuite) TestURLPolicyAllowsListedURLs() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Runner.RunnerAPIURL = "https://runner.example.com"
	cfg.Runner.InitScriptURL = "https://raw.githubusercontent.com/acme/runner/main/init.sh"
	cfg.Runner.PublicAPIURL = "https://anywhere.test"

	ctx, testApp, _ := s.syncEmpty()
	appCfg := s.syncInto(ctx, testApp.ID, cfg, WithURLPolicy(s.urlPolicy(urlpolicy.ModeEnforce)))

	var runnerCfg app.AppRunnerConfig
	s.Require().NoError(s.deps.DB.Where(&app.AppRunnerConfig{AppConfigID: appCfg.ID}).First(&runnerCfg).Error)
	s.Equal("https://runner.example.com", runnerCfg.RunnerAPIURL)
	s.Equal("https://anywhere.test", runnerCfg.PublicAPIURL)
}

func (s *SyncFieldsTestSuite) TestURLPolicyRejectsRunnerURLs() {
	tests := map[string]func(*config.AppConfig){
		"runner.runner_api_url": func(cfg *config.AppConfig) {
			cfg.Runner.RunnerAPIURL = "https://evil.test"
		},
		"runner.init_script_url": func(cfg *config.AppConfig) {
			cfg.Runner.InitScriptURL = "https://raw.githubusercontent.com/mallory/runner/init.sh"
		},
		"runner.env_vars.HTTPS_PROXY": func(cfg *config.AppConfig) {
			cfg.Runner.EnvVarMap = map[string]string{"HTTPS_PROXY": "http://evil.test:3128"}
		},
	}
	for field, mutate := range tests {
		s.Run(field, func() {
			cfg := testseedconfig.BuildMinimalAppConfig()
			mutate(cfg)

			ctx, testApp, _ := s.syncEmpty()
			appCfg, err := s.trySync(ctx, testApp.ID, cfg, WithURLPolicy(s.urlPolicy(urlpolicy.ModeEnforce)))
			s.Require().Error(err)

			var syncErr pkgsync.SyncErr
			s.Require().True(errors.As(err, &syncErr), "a policy violation is a user-facing sync error, got %T", err)
			s.Equal("app-runner", syncErr.Resource)
			s.Contains(syncErr.Description, field)
			s.Zero(s.runnerConfigCount(appCfg.ID), "the sync transaction must roll back")
		})
	}
}

func (s *SyncFieldsTestSuite) TestURLPolicyRejectsStackTemplateURL() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Stack = &config.StackConfig{
		Type:                 string(app.StackTypeAWS),
		Name:                 "stack",
		Description:          "stack",
		VPCNestedTemplateURL: "https://mallory.s3.us-west-2.amazonaws.com/vpc.yaml",
	}

	ctx, testApp, _ := s.syncEmpty()
	_, err := s.trySync(ctx, testApp.ID, cfg, WithURLPolicy(s.urlPolicy(urlpolicy.ModeEnforce)))
	s.Require().Error(err)

	var syncErr pkgsync.SyncErr
	s.Require().True(errors.As(err, &syncErr))
	s.Contains(syncErr.Description, "stack.vpc_nested_template_url")
	s.Contains(syncErr.Description, "allowed_hosts_stack_templates")
}

func (s *SyncFieldsTestSuite) TestURLPolicyWarnModeSyncsAndReports() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Runner.RunnerAPIURL = "https://evil.test"

	policy := s.urlPolicy(urlpolicy.ModeWarn)
	var reported urlpolicy.Violations
	policy.OnWarn(func(v urlpolicy.Violations) { reported = append(reported, v...) })

	ctx, testApp, _ := s.syncEmpty()
	appCfg := s.syncInto(ctx, testApp.ID, cfg, WithURLPolicy(policy))

	s.EqualValues(1, s.runnerConfigCount(appCfg.ID))
	s.Require().Len(reported, 1)
	s.Equal("runner.runner_api_url", reported[0].Field)
}

func (s *SyncFieldsTestSuite) TestURLPolicyUnsetRestrictsNothing() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Runner.RunnerAPIURL = "https://evil.test"
	cfg.Runner.EnvVarMap = map[string]string{"HTTPS_PROXY": "http://evil.test:3128"}

	ctx, testApp, _ := s.syncEmpty()
	appCfg := s.syncInto(ctx, testApp.ID, cfg)
	s.EqualValues(1, s.runnerConfigCount(appCfg.ID))
}
