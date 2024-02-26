package config

import "fmt"

// Component is a flattened configuration type that allows us to define components using a `type: type` field.
type Component struct {
	Name string `mapstructure:"name" toml:"name"`
	Type string `mapstructure:"type" toml:"type"`

	Data map[string]interface{} `mapstructure:",remain"`

	TerraformModule *TerraformModuleComponentConfig `mapstructure:"-"`
	HelmChart       *HelmChartComponentConfig       `mapstructure:"-"`
	DockerBuild     *DockerBuildComponentConfig     `mapstructure:"-"`
	ExternalImage   *ExternalImageComponentConfig   `mapstructure:"-"`
	Job             *JobComponentConfig             `mapstructure:"-"`
}

func (c *Component) ToResourceType() string {
	if c.TerraformModule != nil {
		return "nuon_terraform_module_component"
	}
	if c.HelmChart != nil {
		return "nuon_helm_chart_component"
	}
	if c.DockerBuild != nil {
		return "nuon_docker_build_component"
	}
	if c.ExternalImage != nil {
		return "nuon_external_image_component"
	}
	if c.Job != nil {
		return "nuon_job_coponent"
	}

	return ""
}

func (c *Component) ToResource() (map[string]interface{}, error) {
	var (
		resource map[string]interface{}
		err      error
	)

	if c.TerraformModule != nil {
		resource, err = c.TerraformModule.ToResource()
	}
	if c.HelmChart != nil {
		resource, err = c.HelmChart.ToResource()
	}
	if c.DockerBuild != nil {
		resource, err = c.DockerBuild.ToResource()
	}
	if c.ExternalImage != nil {
		resource, err = c.ExternalImage.ToResource()
	}
	if c.Job != nil {
		resource, err = c.Job.ToResource()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to generate resource: %w", err)
	}
	if resource == nil {
		return nil, fmt.Errorf("invalid component type")
	}

	resource["name"] = c.Name

	return nestWithName(c.Name, resource), nil
}
