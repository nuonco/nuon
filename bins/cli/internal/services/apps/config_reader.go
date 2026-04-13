package apps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/generator"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/pkg/errors"
)

const (
	ExampleAppConfigsRepo = "https://github.com/nuonco/example-app-configs"
	DefaultBranch         = "main"
	DefaultTempDirectory  = "/tmp/"
)

// ConfigReaderParams holds parameters for reading remote app configs
type ConfigReaderParams struct {
	// RepoURL url to configs repo, default nuonco/example-app-config
	RepoURL string
	// Branch, default branch main
	Branch string
	// Folder is app config folder
	Folder string
	// TempDir is location of temp directory used to clone repo
	TempDir string
}

// ConfigReader handles cloning and reading app configs from remote repositories
type ConfigReader struct {
	params    ConfigReaderParams
	clonePath string
}

func NewConfigReader(params ConfigReaderParams) (*ConfigReader, error) {
	// defaults
	if params.RepoURL == "" {
		params.RepoURL = ExampleAppConfigsRepo
	}
	if params.Branch == "" {
		params.Branch = DefaultBranch
	}
	if params.TempDir == "" {
		params.TempDir = DefaultTempDirectory
	}

	// validation
	if params.Folder == "" {
		return nil, errors.New("folder parameter is required")
	}

	return &ConfigReader{
		params: params,
	}, nil
}

func (r *ConfigReader) Clone(ctx context.Context) error {
	if r.params.TempDir == "" {
		tempDir, err := os.MkdirTemp("", "nuon-example-configs-*")
		if err != nil {
			return errors.Wrap(err, "failed to create temporary directory")
		}
		r.params.TempDir = tempDir
	}

	r.clonePath = filepath.Join(r.params.TempDir, "repo")

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", r.params.Branch, r.params.RepoURL, r.clonePath, "--quiet")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to clone repository %s (branch: %s)", r.params.RepoURL, r.params.Branch)
	}

	return nil
}

func (r *ConfigReader) ReadConfig(ctx context.Context) (*config.AppConfig, error) {
	if r.clonePath == "" {
		return nil, errors.New("repository not cloned yet, call Clone() first")
	}

	configPath := filepath.Join(r.clonePath, r.params.Folder)

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("config folder '%s' does not exist in repository", r.params.Folder)
		}
		return nil, errors.Wrap(err, "failed to access config folder")
	}

	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:       configPath,
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to read app config")
	}

	return cfg, nil
}

func (r *ConfigReader) Cleanup() error {
	if r.params.TempDir != "" {
		if err := os.RemoveAll(r.clonePath); err != nil {
			return errors.Wrap(err, "failed to cleanup temporary directory")
		}
	}
	return nil
}

func ConvertToConfigStructure(appConfig *config.AppConfig, targetPath string) (*generator.ConfigStructure, error) {
	if appConfig == nil {
		return nil, errors.New("appConfig cannot be nil")
	}

	structure := &generator.ConfigStructure{
		Name:              targetPath,
		Configs:           []generator.ConfigFileDefinition{},
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	if appConfig.Version != "" || appConfig.DisplayName != "" || appConfig.Description != "" {
		metadataConfig := &config.MetadataConfig{
			Version:         appConfig.Version,
			DisplayName:     appConfig.DisplayName,
			Description:     appConfig.Description,
			SlackWebhookURL: appConfig.SlackWebhookURL,
			Readme:          appConfig.Readme,
		}
		if err := structure.UpdateMetadata(metadataConfig); err != nil {
			return nil, errors.Wrap(err, "failed to update metadata")
		}
	}

	if appConfig.Inputs != nil {
		if err := structure.UpdateInputs(appConfig.Inputs); err != nil {
			return nil, errors.Wrap(err, "failed to update inputs")
		}
	}

	if appConfig.Sandbox != nil {
		if err := structure.UpdateSandbox(appConfig.Sandbox); err != nil {
			return nil, errors.Wrap(err, "failed to update sandbox")
		}
	}

	if appConfig.Stack != nil {
		if err := structure.UpdateStack(appConfig.Stack); err != nil {
			return nil, errors.Wrap(err, "failed to update stack")
		}
	}

	if appConfig.Runner != nil {
		if err := structure.UpdateRunner(appConfig.Runner); err != nil {
			return nil, errors.Wrap(err, "failed to update runner")
		}
	}

	if appConfig.Secrets != nil {
		if err := structure.UpdateSecrets(appConfig.Secrets); err != nil {
			return nil, errors.Wrap(err, "failed to update secrets")
		}
	}

	if appConfig.BreakGlass != nil {
		if err := structure.UpdateBreakGlass(appConfig.BreakGlass); err != nil {
			return nil, errors.Wrap(err, "failed to update break glass")
		}
	}

	if appConfig.Policies != nil {
		if err := structure.UpdatePolicies(appConfig.Policies); err != nil {
			return nil, errors.Wrap(err, "failed to update policies")
		}
	}

	for _, component := range appConfig.Components {
		if component == nil {
			continue
		}

		var componentConfig any
		switch component.Type {
		case config.ComponentType(config.TerraformModuleComponentType):
			componentConfig = component.TerraformModule
			component.TerraformModule = nil
		case config.ComponentType(config.KubernetesManifestComponentType):
			componentConfig = component.KubernetesManifest
			component.KubernetesManifest = nil
		case config.ComponentType(config.DockerBuildComponentType):
			componentConfig = component.DockerBuild
			component.DockerBuild = nil
		case config.ComponentType(config.HelmChartComponentType):
			componentConfig = component.HelmChart
			component.HelmChart = nil
		case config.ComponentType(config.ContainerImageComponentType):
			componentConfig = component.ExternalImage
			component.ExternalImage = nil
		case config.ComponentType(config.JobComponentType):
			componentConfig = component.Job
			component.Job = nil
		case config.ComponentType(config.PulumiComponentType):
			componentConfig = component.Pulumi
			component.Pulumi = nil
		}

		componentDef := generator.ConfigFileDefinition{
			Name: component.Name + ".toml",
			Schemas: []generator.ConfigFileSchema{
				{
					Instance: component,
				},
				{
					Instance: componentConfig,
				},
			},
		}

		if err := structure.AddComponent(componentDef); err != nil {
			return nil, errors.Wrapf(err, "failed to add component '%s'", component.Name)
		}
	}

	for _, action := range appConfig.Actions {
		if action == nil {
			continue
		}

		actionName := action.Name
		if actionName == "" {
			actionName = "action"
		}

		actionDef := generator.ConfigFileDefinition{
			Name: actionName + ".toml",
			Schemas: []generator.ConfigFileSchema{
				{
					Instance: action,
				},
			},
		}

		if err := structure.AddActions(actionDef); err != nil {
			return nil, errors.Wrapf(err, "failed to add action '%s'", actionName)
		}
	}

	if appConfig.Permissions != nil {
		for _, role := range appConfig.Permissions.Roles {
			if role == nil {
				continue
			}

			roleName := role.Type

			permissionDef := generator.ConfigFileDefinition{
				Name: roleName + ".toml",
				Schemas: []generator.ConfigFileSchema{
					{
						Instance: role,
					},
				},
			}

			if err := structure.AddPermission(permissionDef); err != nil {
				return nil, errors.Wrapf(err, "failed to add permission '%s'", roleName)
			}
		}
	}

	return structure, nil
}

// ReadAndConvertConfig is a convenience function that clones, reads, and converts the config in one call
func ReadAndConvertConfig(ctx context.Context, params ConfigReaderParams, targetPath string) (*generator.ConfigStructure, error) {
	reader, err := NewConfigReader(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create config reader")
	}

	// cleanup happens
	defer func() {
		if err := reader.Cleanup(); err != nil {
			fmt.Println(errors.Wrap(err, "Warning: failed to cleanup temporary directory: %v\n"))
		}
	}()

	// clone
	if err := reader.Clone(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to clone repository")
	}

	// parse config
	appConfig, err := reader.ReadConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	// convert to config structure
	configStructure, err := ConvertToConfigStructure(appConfig, targetPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert to config structure")
	}

	return configStructure, nil
}
