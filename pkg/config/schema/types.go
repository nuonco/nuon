package schema

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config"
)

func LookupSchemaType(typ string) (*jsonschema.Schema, error) {
	mapping := map[string]func() (*jsonschema.Schema, error){
		"full":                AppConfigSchema,
		"runner":              RunnerConfigSchema,
		"sandbox":             SandboxConfigSchema,
		"helm":                HelmConfigSchema,
		"docker-build":        DockerBuildConfigSchema,
		"kubernetes-manifest": KubernetesManifestConfigSchema,
		"terraform":           TerraformModuleConfigSchema,
		"container-image":     ContainerImageConfigSchema,
		"permissions":         PermissionsConfigSchema,
		"policy":              PolicyConfigSchema,
		"secrets":             SecretsConfigSchema,
		"secret":              SecretConfigSchema,
		"metadata":            MetadataConfigSchema,
		"action":              ActionConfigSchema,
		"stack":               StackConfigSchema,
		"installer":           InstallerConfigSchema,
		"break-glass":         BreakGlassConfigSchema,
		"inputs":              InputsConfigSchema,
		"input-group":         InputGroupSchema,
		"input":               InputSchema,
		"install":             InstallSchema,
	}

	fn, ok := mapping[typ]
	if !ok {
		return nil, fmt.Errorf("no schema found for type %s", typ)
	}

	schema, err := fn()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get schema")
	}

	return schema, nil
}

func InputSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInput{}), nil
}

func InputGroupSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInputGroup{}), nil
}

func InputsConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInputConfig{}), nil
}

func AppConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppConfig{}), nil
}

func RunnerConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppRunnerConfig{}), nil
}

func SandboxConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppSandboxConfig{}), nil
}

func HelmConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.HelmChartComponentConfig{}), nil
}

func KubernetesManifestConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.KubernetesManifestComponentConfig{}), nil
}

func DockerBuildConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.DockerBuildComponentConfig{}), nil
}

func TerraformModuleConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.TerraformModuleComponentConfig{}), nil
}

func ContainerImageConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.ExternalImageComponentConfig{}), nil
}

func PermissionsConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.PermissionsConfig{}), nil
}

func PolicyConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppPolicy{}), nil
}

func SecretsConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.SecretsConfig{}), nil
}

func SecretConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppSecret{}), nil
}

func MetadataConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.MetadataConfig{}), nil
}

func ActionConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.ActionConfig{}), nil
}

func StackConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.StackConfig{}), nil
}

func InstallerConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.InstallerConfig{}), nil
}

func BreakGlassConfigSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppAWSIAMRole{}), nil
}

func InstallSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.Install{}), nil
}
