package schema

import (
	"github.com/invopop/jsonschema"

	"github.com/powertoolsdev/mono/pkg/config"
)

func SandboxSourceSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppSandboxConfig{}), nil
}

func InstallerSourceSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.InstallerConfig{}), nil
}

func RunnerSourceSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppRunnerConfig{}), nil
}

func InputsSourceSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInputConfig{}), nil
}

func DockerBuildComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.DockerBuildComponentConfig{}), nil
}

func HelmComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.HelmChartComponentConfig{}), nil
}

func ContainerImageComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.ExternalImageComponentConfig{}), nil
}

func TerraformComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.TerraformModuleComponentConfig{}), nil
}

func JobComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.JobComponentConfig{}), nil
}
