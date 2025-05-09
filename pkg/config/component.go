package config

import (
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type ComponentType string

const (
	// TerraformModuleComponentType is the type for a terraform module component
	TerraformModuleComponentType ComponentType = "terraform_module"
	// HelmChartComponentType is the type for a helm chart component
	HelmChartComponentType ComponentType = "helm_chart"
	// DockerBuildComponentType is the type for a docker build component
	DockerBuildComponentType ComponentType = "docker_build"
	// ContainerImageComponentType is the type for an external image component
	ContainerImageComponentType ComponentType = "container_image"
	ExternalImageComponentType  ComponentType = "external_image"
	// JobComponentType is the type for a job component
	JobComponentType ComponentType = "job"

	ComponentTypeUnknown ComponentType = ""
)

func (c ComponentType) APIType() models.AppComponentType {
	switch c {
	case TerraformModuleComponentType:
		return models.AppComponentTypeTerraformModule
	case HelmChartComponentType:
		return models.AppComponentTypeHelmChart
	case DockerBuildComponentType:
		return models.AppComponentTypeDockerBuild
	case ContainerImageComponentType:
		return models.AppComponentTypeExternalImage
	case JobComponentType:
		return models.AppComponentTypeJob
	}

	return models.AppComponentTypeUnknown
}

// Component is a flattened configuration type that allows us to define components using a `type: type` field.
type Component struct {
	Source string `mapstructure:"source,omitempty"`

	Type         ComponentType `mapstructure:"type,omitempty" jsonschema:"required"`
	Name         string        `mapstructure:"name" jsonschema:"required"`
	VarName      string        `mapstructure:"var_name,omitempty"`
	Dependencies []string      `mapstructure:"dependencies,omitempty"`

	HelmChart       *HelmChartComponentConfig       `mapstructure:"helm_chart,omitempty" jsonschema:"oneof_required=helm"`
	TerraformModule *TerraformModuleComponentConfig `mapstructure:"terraform_module,omitempty" jsonschema:"oneof_required=terraform_module"`
	DockerBuild     *DockerBuildComponentConfig     `mapstructure:"docker_build,omitempty" jsonschema:"oneof_required=docker_build"`
	Job             *JobComponentConfig             `mapstructure:"job,omitempty" jsonschema:"oneof_required=job"`
	ExternalImage   *ExternalImageComponentConfig   `mapstructure:"external_image,omitempty" jsonschema:"oneof_required=external_image"`

	// created during parsing
	References []refs.Ref `mapstructure:"-" jsonschema:"-"`
}

func (c *Component) parse() error {
	references, err := refs.Parse(c)
	if err != nil {
		return errors.Wrap(err, "unable to parse components")
	}
	c.References = references

	// set all of the components
	for _, ref := range c.References {
		if !generics.SliceContains(ref.Type, []refs.RefType{refs.RefTypeComponents, refs.RefTypeComponentsNested}) {
			continue
		}

		c.Dependencies = append(c.Dependencies, ref.Name)
	}
	c.Dependencies = generics.UniqueSlice(c.Dependencies)

	if c.HelmChart != nil {
		return c.HelmChart.Parse()
	}

	if c.TerraformModule != nil {
		return c.TerraformModule.Parse()
	}

	if c.DockerBuild != nil {
		return c.DockerBuild.Parse()
	}

	if c.Job != nil {
		return c.Job.Parse()
	}

	if c.ExternalImage != nil {
		return c.ExternalImage.Parse()
	}

	return nil
}

func (c *Component) AddDependency(val string) {
	c.Dependencies = append(c.Dependencies, val)
}

func (c *Component) AllVars() []string {
	vars := make([]string, 0)

	if c.HelmChart != nil {
		for _, v := range c.HelmChart.Values {
			vars = append(vars, v.Value)
		}
		for _, v := range c.HelmChart.ValuesMap {
			vars = append(vars, v)
		}
	}
	if c.TerraformModule != nil {
		for _, v := range c.TerraformModule.Variables {
			vars = append(vars, v.Value)
		}

		for _, v := range c.TerraformModule.EnvVars {
			vars = append(vars, v.Value)
		}

		for _, v := range c.TerraformModule.EnvVarMap {
			vars = append(vars, v)
		}
	}

	return vars
}
