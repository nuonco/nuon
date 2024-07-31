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

	schema := r.Reflect(config.DockerBuildComponentConfig{})
	compSchema := r.Reflect(&config.Component{})

	typProp, _ := compSchema.Definitions["Component"].Properties.Get("type")
	schema.Definitions["DockerBuildComponentConfig"].Properties.Set("type", typProp)

	nameProp, _ := compSchema.Definitions["Component"].Properties.Get("name")
	schema.Definitions["DockerBuildComponentConfig"].Properties.Set("name", nameProp)
	schema.Definitions["DockerBuildComponentConfig"].Required = append(schema.Definitions["DockerBuildComponentConfig"].Required, "name")
	schema.Definitions["DockerBuildComponentConfig"].Required = append(schema.Definitions["DockerBuildComponentConfig"].Required, "type")

	return schema, nil
}

func HelmComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	schema := r.Reflect(config.HelmChartComponentConfig{})
	compSchema := r.Reflect(config.Component{})

	typProp, _ := compSchema.Definitions["Component"].Properties.Get("type")
	schema.Definitions["HelmChartComponentConfig"].Properties.Set("type", typProp)

	nameProp, _ := compSchema.Definitions["Component"].Properties.Get("name")
	schema.Definitions["HelmChartComponentConfig"].Properties.Set("name", nameProp)
	schema.Definitions["HelmChartComponentConfig"].Required = append(schema.Definitions["HelmChartComponentConfig"].Required, "name")
	schema.Definitions["HelmChartComponentConfig"].Required = append(schema.Definitions["HelmChartComponentConfig"].Required, "type")

	return schema, nil
}

func ExternalImageComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	schema := r.Reflect(config.ExternalImageComponentConfig{})
	compSchema := r.Reflect(config.Component{})

	typProp, _ := compSchema.Definitions["Component"].Properties.Get("type")
	schema.Definitions["ExternalImageComponentConfig"].Properties.Set("type", typProp)

	nameProp, _ := compSchema.Definitions["Component"].Properties.Get("name")
	schema.Definitions["ExternalImageComponentConfig"].Properties.Set("name", nameProp)
	schema.Definitions["ExternalImageComponentConfig"].Required = append(schema.Definitions["ExternalImageComponentConfig"].Required, "name")
	schema.Definitions["ExternalImageComponentConfig"].Required = append(schema.Definitions["ExternalImageComponentConfig"].Required, "type")

	return schema, nil
}

func TerraformComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	schema := r.Reflect(config.TerraformModuleComponentConfig{})
	compSchema := r.Reflect(config.Component{})

	typProp, _ := compSchema.Definitions["Component"].Properties.Get("type")
	schema.Definitions["TerraformModuleComponentConfig"].Properties.Set("type", typProp)

	nameProp, _ := compSchema.Definitions["Component"].Properties.Get("name")
	schema.Definitions["TerraformModuleComponentConfig"].Properties.Set("name", nameProp)
	schema.Definitions["TerraformModuleComponentConfig"].Required = append(schema.Definitions["TerraformModuleComponentConfig"].Required, "name")
	schema.Definitions["TerraformModuleComponentConfig"].Required = append(schema.Definitions["TerraformModuleComponentConfig"].Required, "type")

	return schema, nil
}

func JobComponent() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	schema := r.Reflect(config.JobComponentConfig{})
	compSchema := r.Reflect(config.Component{})

	typProp, _ := compSchema.Definitions["Component"].Properties.Get("type")
	schema.Definitions["JobComponentConfig"].Properties.Set("type", typProp)

	nameProp, _ := compSchema.Definitions["Component"].Properties.Get("name")
	schema.Definitions["JobComponentConfig"].Properties.Set("name", nameProp)
	schema.Definitions["JobComponentConfig"].Required = append(schema.Definitions["JobComponentConfig"].Required, "name")
	schema.Definitions["JobComponentConfig"].Required = append(schema.Definitions["JobComponentConfig"].Required, "type")

	return schema, nil
}
