package sync

import (
	"context"
	"reflect"
	"sort"

	"github.com/mitchellh/copystructure"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/config"
)

type sync struct {
	cfg       *config.AppConfig
	apiClient nuon.Client
	appID     string
}

func (s sync) syncApp(ctx context.Context) error {
	currentApp, err := s.apiClient.GetApp(ctx, s.appID)
	if err != nil {
		return err
	}

	eq := s.compareApp(currentApp)
	if eq {
		return nil
	}
	_, err = s.apiClient.UpdateApp(ctx, s.appID, &models.ServiceUpdateAppRequest{
		Description:     s.cfg.Description,
		DisplayName:     s.cfg.DisplayName,
		SlackWebhookURL: s.cfg.SlackWebhookURL,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s sync) syncAppSandbox(ctx context.Context) error {
	currentSandBoxConfig, err := s.apiClient.GetAppSandboxLatestConfig(ctx, s.appID)
	if err != nil {
		return err
	}

	eq, err := s.compareAppSandbox(currentSandBoxConfig)
	if err != nil {
		return err
	}
	if eq {
		return nil
	}

	req := &models.ServiceCreateAppSandboxConfigRequest{
		SandboxInputs:    s.cfg.Sandbox.VarMap,
		TerraformVersion: &s.cfg.Sandbox.TerraformVersion,
	}

	if s.cfg.Sandbox.ConnectedRepo != nil {
		req.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSSandboxConfigRequest{
			Repo:      &s.cfg.Sandbox.ConnectedRepo.Repo,
			Branch:    s.cfg.Sandbox.ConnectedRepo.Branch,
			Directory: &s.cfg.Sandbox.ConnectedRepo.Directory,
		}
	}

	if s.cfg.Sandbox.PublicRepo != nil {
		req.PublicGitVcsConfig = &models.ServicePublicGitVCSSandboxConfigRequest{
			Repo:      &s.cfg.Sandbox.PublicRepo.Repo,
			Branch:    &s.cfg.Sandbox.PublicRepo.Branch,
			Directory: &s.cfg.Sandbox.PublicRepo.Directory,
		}
	}

	_, err = s.apiClient.CreateAppSandboxConfig(ctx, s.appID, req)

	if err != nil {
		return err
	}

	return nil
}

func (s sync) syncAppRunner(ctx context.Context) error {
	currentRunnerConfigs, err := s.apiClient.GetAppRunnerLatestConfig(ctx, s.appID)
	if err != nil {
		return err
	}

	eq := s.compareAppRunner(currentRunnerConfigs)
	if eq {
		return nil
	}
	newCfgEnvVars := make(map[string]string)
	for _, v := range s.cfg.Runner.EnvironmentVariables {
		newCfgEnvVars[v.Name] = v.Value
	}
	_, err = s.apiClient.CreateAppRunnerConfig(ctx, s.appID, &models.ServiceCreateAppRunnerConfigRequest{
		EnvVars: s.cfg.Runner.EnvVarMap,
		Type:    models.AppAppRunnerType(s.cfg.Runner.RunnerType),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s sync) syncAppInput(ctx context.Context) error {
	// NOTE: if the cli has previously failed to initially load inputs, it can return a 404
	currentAppInputs, err := s.apiClient.GetAppInputLatestConfig(ctx, s.appID)
	if err != nil && !nuon.IsNotFound(err) {
		return err
	}

	if err == nil && !nuon.IsNotFound(err) {
		eq, err := s.compareAppInputs(currentAppInputs)
		if err != nil {
			return err
		}
		if eq {
			return nil
		}
	}

	groups := make(map[string]models.ServiceAppGroupRequest)
	for _, group := range s.cfg.Inputs.Groups {
		group := group
		groups[group.Name] = models.ServiceAppGroupRequest{
			Description: &group.Description,
			DisplayName: &group.DisplayName,
		}
		newGroup := models.ServiceAppGroupRequest{}
		newGroup.Description = &group.Description
		newGroup.DisplayName = &group.DisplayName
		groups[group.Name] = newGroup
	}

	inputs := make(map[string]models.ServiceAppInputRequest)
	for _, input := range s.cfg.Inputs.Inputs {
		input := input
		inputs[input.Name] = models.ServiceAppInputRequest{
			Description: &input.Description,
			DisplayName: &input.DisplayName,
			Group:       &input.Group,
			Required:    input.Required,
			Sensitive:   input.Sensitive,
		}
	}

	_, err = s.apiClient.CreateAppInputConfig(ctx, s.appID, &models.ServiceCreateAppInputConfigRequest{
		Groups: groups,
		Inputs: inputs,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *sync) Sync(ctx context.Context) error {
	if err := s.syncApp(ctx); err != nil {
		return err
	}

	if err := s.syncAppSandbox(ctx); err != nil {
		return err
	}

	if err := s.syncAppRunner(ctx); err != nil {
		return err
	}

	if err := s.syncAppInput(ctx); err != nil {
		return err
	}

	// TODO sync installer

	// TODO sync components

	return nil
}

func (s *sync) compareApp(currentApp *models.AppApp) bool {
	// cfg does not have name field so we ignore it
	// *models.AppApp does not include SlackWebUrl we may need to fix to support syncing that property
	if currentApp.Description == s.cfg.Description &&
		currentApp.DisplayName == s.cfg.DisplayName {
		return true
	}

	return false
}

func (s *sync) compareAppSandbox(currentSandBoxConfig *models.AppAppSandboxConfig) (bool, error) {
	i, err := copystructure.Copy(s.cfg.Sandbox)
	currentCfg := i.(*config.AppSandboxConfig)
	if err != nil {
		return false, err
	}

	if currentSandBoxConfig.ConnectedGithubVcsConfig != nil {
		currentCfg.ConnectedRepo = &config.ConnectedRepoConfig{
			Repo:      currentCfg.ConnectedRepo.Repo,
			Branch:    currentCfg.ConnectedRepo.Branch,
			Directory: currentCfg.ConnectedRepo.Directory,
		}
	}

	if currentSandBoxConfig.PublicGitVcsConfig != nil {
		currentCfg.PublicRepo = &config.PublicRepoConfig{
			Repo:      currentCfg.PublicRepo.Repo,
			Branch:    currentCfg.PublicRepo.Branch,
			Directory: currentCfg.PublicRepo.Directory,
		}
	}

	currentCfgVars := make([]config.TerraformVariable, len(currentSandBoxConfig.Variables))
	idx := 0
	for k, v := range currentSandBoxConfig.Variables {
		currentCfgVars[idx] = config.TerraformVariable{
			Name:  k,
			Value: v,
		}
		idx++
	}
	currentCfg.Vars = currentCfgVars
	currentCfg.TerraformVersion = currentSandBoxConfig.TerraformVersion

	// sort slices for comparison
	sort.Slice(currentCfg.Vars, func(i, j int) bool {
		return currentCfg.Vars[i].Name < currentCfg.Vars[j].Name
	})
	sort.Slice(s.cfg.Sandbox.Vars, func(i, j int) bool {
		return s.cfg.Sandbox.Vars[i].Name < s.cfg.Sandbox.Vars[j].Name
	})

	if reflect.DeepEqual(s.cfg.Sandbox, currentCfg) {
		return true, nil
	}

	return false, nil
}

func (s *sync) compareAppRunner(currentApp *models.AppAppRunnerConfig) bool {
	newCfgEnvVars := make(map[string]string)
	for _, v := range s.cfg.Runner.EnvironmentVariables {
		newCfgEnvVars[v.Name] = v.Value
	}
	if currentApp.AppRunnerType == models.AppAppRunnerType(s.cfg.Runner.RunnerType) &&
		reflect.DeepEqual(currentApp.EnvVars, newCfgEnvVars) {
		return true
	}
	return false
}

func (s *sync) compareAppInputs(currentApp *models.AppAppInputConfig) (bool, error) {
	if currentApp != nil && s.cfg.Inputs == nil {
		return false, nil
	}

	if currentApp == nil && s.cfg.Inputs != nil {
		return false, nil
	}

	i, err := copystructure.Copy(s.cfg.Inputs)
	cfgCurrent := i.(*config.AppInputConfig)

	if err != nil {
		return false, err
	}

	groups := make([]config.AppInputGroup, len(currentApp.InputGroups))
	for idx, group := range currentApp.InputGroups {
		groups[idx] = config.AppInputGroup{
			Name:        group.Name,
			Description: group.Description,
			DisplayName: group.DisplayName,
		}
	}

	cfgCurrent.Groups = groups

	inputs := make([]config.AppInput, len(currentApp.Inputs))
	for idx, input := range currentApp.Inputs {
		inputs[idx] = config.AppInput{
			Name:        input.Name,
			Description: input.Description,
			DisplayName: input.DisplayName,
			Group:       input.Group.Name,
			Required:    input.Required,
			Sensitive:   input.Sensitive,
		}
	}
	cfgCurrent.Inputs = inputs

	// sort slices for comparison
	sort.Slice(cfgCurrent.Inputs, func(i, j int) bool {
		return cfgCurrent.Inputs[i].Name < cfgCurrent.Inputs[j].Name
	})
	sort.Slice(s.cfg.Inputs.Inputs, func(i, j int) bool {
		return s.cfg.Inputs.Inputs[i].Name < s.cfg.Inputs.Inputs[j].Name
	})
	sort.Slice(cfgCurrent.Groups, func(i, j int) bool {
		return cfgCurrent.Groups[i].Name < cfgCurrent.Groups[j].Name
	})
	sort.Slice(s.cfg.Inputs.Groups, func(i, j int) bool {
		return s.cfg.Inputs.Groups[i].Name < s.cfg.Inputs.Groups[j].Name
	})

	// WARN: slice properties can be out of order causing a false negative
	if reflect.DeepEqual(s.cfg.Inputs, cfgCurrent) {
		return true, nil
	}

	return false, nil
}

func New(apiClient nuon.Client, appID string, cfg *config.AppConfig) *sync {
	return &sync{
		cfg:       cfg,
		apiClient: apiClient,
		appID:     appID,
	}
}
