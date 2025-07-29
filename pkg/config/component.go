package config

import (
	"sort"

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
	// KubernetesManifestComponentType is a type for kubernetes manifest compnent
	KubernetesManifestComponentType ComponentType = "kubernetes_manifest"

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
	case KubernetesManifestComponentType:
		return models.AppComponentTypeKubernetesManifest
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

	HelmChart          *HelmChartComponentConfig          `mapstructure:"helm_chart,omitempty" jsonschema:"oneof_required=helm"`
	TerraformModule    *TerraformModuleComponentConfig    `mapstructure:"terraform_module,omitempty" jsonschema:"oneof_required=terraform_module"`
	DockerBuild        *DockerBuildComponentConfig        `mapstructure:"docker_build,omitempty" jsonschema:"oneof_required=docker_build"`
	Job                *JobComponentConfig                `mapstructure:"job,omitempty" jsonschema:"oneof_required=job"`
	ExternalImage      *ExternalImageComponentConfig      `mapstructure:"external_image,omitempty" jsonschema:"oneof_required=external_image"`
	KubernetesManifest *KubernetesManifestComponentConfig `mapstructure:"kubernetes_manifest,omitempty" jsonschema:"oneof_required=kubernetes_manifest"`

	// created during parsing
	References []refs.Ref `mapstructure:"-" jsonschema:"-" nuonhash:"-"`
	Checksum   string     `mapstructure:"-" jsonschema:"-" toml:"checksum" nuonhash:"-"`
}

func (c *Component) parse() error {
	if c == nil {
		return nil
	}

	if c.HelmChart != nil {
		if err := c.HelmChart.Parse(); err != nil {
			return err
		}
	}

	if c.TerraformModule != nil {
		if err := c.TerraformModule.Parse(); err != nil {
			return err
		}
	}

	if c.DockerBuild != nil {
		if err := c.DockerBuild.Parse(); err != nil {
			return err
		}
	}

	if c.ExternalImage != nil {
		if err := c.ExternalImage.Parse(); err != nil {
			return err
		}
	}

	if c.KubernetesManifest != nil {
		if err := c.KubernetesManifest.Parse(); err != nil {
			return err
		}
	}

	references, err := refs.Parse(c)
	if err != nil {
		return errors.Wrap(err, "unable to parse components")
	}
	c.References = references

	// set all of the components
	for _, ref := range c.References {
		if !generics.SliceContains(ref.Type, []refs.RefType{refs.RefTypeComponents}) {
			continue
		}

		c.Dependencies = append(c.Dependencies, ref.Name)
	}
	c.Dependencies = generics.UniqueSlice(c.Dependencies)
	sort.Strings(c.Dependencies)

	return nil
}

func (a *Component) Validate() error {
	if a.HelmChart != nil {
		return a.HelmChart.Validate()
	}

	if a.TerraformModule != nil {
		return a.TerraformModule.Validate()
	}

	if a.DockerBuild != nil {
		return a.DockerBuild.Validate()
	}

	if a.ExternalImage != nil {
		return a.ExternalImage.Validate()
	}

	if a.KubernetesManifest != nil {
		return a.KubernetesManifest.Validate()
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
