package config

import (
	"sort"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
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
	Source string `mapstructure:"source,omitempty" toml:"source,omitempty"`

	Type         ComponentType `mapstructure:"type,omitempty" toml:"type,omitempty" jsonschema:"required"`
	Name         string        `mapstructure:"name" toml:"name" jsonschema:"required"`
	VarName      string        `mapstructure:"var_name,omitempty" toml:"var_name,omitempty"`
	Dependencies []string      `mapstructure:"dependencies,omitempty" toml:"dependencies,omitempty"`

	// WARNING: properties below should be ignored by nuonhash when empty
	HelmChart          *HelmChartComponentConfig          `mapstructure:"helm_chart,omitempty" toml:"helm_chart,omitempty" jsonschema:"oneof_required=helm" nuonhash:"omitempty"`
	TerraformModule    *TerraformModuleComponentConfig    `mapstructure:"terraform_module,omitempty" toml:"terraform_module,omitempty" jsonschema:"oneof_required=terraform_module" nuonhash:"omitempty"`
	DockerBuild        *DockerBuildComponentConfig        `mapstructure:"docker_build,omitempty" toml:"docker_build,omitempty" jsonschema:"oneof_required=docker_build" nuonhash:"omitempty"`
	Job                *JobComponentConfig                `mapstructure:"job,omitempty" toml:"job,omitempty" jsonschema:"oneof_required=job" nuonhash:"omitempty"`
	ExternalImage      *ExternalImageComponentConfig      `mapstructure:"external_image,omitempty" toml:"external_image,omitempty" jsonschema:"oneof_required=external_image" nuonhash:"omitempty"`
	KubernetesManifest *KubernetesManifestComponentConfig `mapstructure:"kubernetes_manifest,omitempty" toml:"kubernetes_manifest,omitempty" jsonschema:"oneof_required=kubernetes_manifest" nuonhash:"omitempty"`

	// created during parsing
	// WARNING: properties below should not be hashed with nuonhash
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

func (c Component) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("component type").Required().
		Long("Type of component to deploy. Determines which configuration block is required (helm_chart, terraform_module, docker_build, container_image, kubernetes_manifest, or job)").
		Example("terraform_module").
		Example("helm_chart").
		Example("docker_build").
		Example("container_image").
		Example("kubernetes_manifest").
		Field("name").Short("component name").Required().
		Long("Unique identifier for the component within the app. Used for referencing in dependencies and templates").
		Example("database").
		Example("api-server").
		Example("frontend").
		Field("source").Short("source path or URL").
		Long("Optional source path or URL for the component configuration. Supports HTTP(S) URLs, git repositories, file paths, and relative paths (./). Examples: https://example.com/config.yaml, git::https://github.com/org/repo//config.yaml, file:///path/to/config.yaml, ./local/config.yaml").
		Field("var_name").Short("variable name for component output").
		Long("Optional name to use when storing component outputs as variables. If not specified, uses the component name").
		Example("db_endpoint").
		Example("api_host").
		Field("dependencies").Short("component dependencies").
		Long("List of other components that must be deployed before this component. Automatically extracted from template references").
		Example("database").
		Example("infrastructure").
		Field("helm_chart").Short("helm chart component configuration").OneOfRequired("component_type").
		Long("Configuration for Helm chart deployments. Required when type is 'helm_chart'").
		Field("terraform_module").Short("terraform module component configuration").OneOfRequired("component_type").
		Long("Configuration for Terraform module deployments. Required when type is 'terraform_module'").
		Field("docker_build").Short("docker build component configuration").OneOfRequired("component_type").
		Long("Configuration for building and pushing Docker images. Required when type is 'docker_build'").
		Field("external_image").Short("container image component configuration").OneOfRequired("component_type").
		Long("Configuration for external container images (e.g., from Docker Hub or ECR). Required when type is 'container_image' or 'external_image'").
		Field("kubernetes_manifest").Short("kubernetes manifest component configuration").OneOfRequired("component_type").
		Long("Configuration for Kubernetes manifest deployments. Required when type is 'kubernetes_manifest'").
		Field("job").Short("job component configuration").OneOfRequired("component_type").Deprecated("").
		Long("Configuration for job/batch components. Required when type is 'job'")
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
