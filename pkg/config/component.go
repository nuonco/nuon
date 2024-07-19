package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

const (
	// TerraformModuleComponentType is the type for a terraform module component
	TerraformModuleComponentType = "terraform_module"
	// HelmChartComponentType is the type for a helm chart component
	HelmChartComponentType = "helm_chart"
	// DockerBuildComponentType is the type for a docker build component
	DockerBuildComponentType = "docker_build"
	// ContainerImageComponentType is the type for an external image component
	ContainerImageComponentType = "container_image"
	// JobComponentType is the type for a job component
	JobComponentType = "job"
)

// Component is a flattened configuration type that allows us to define components using a `type: type` field.
type Component map[string]interface{}

type MinComponent struct {
	Source  string `mapstructure:"source,omitempty"`
	Name    string `mapstructure:"name,omitempty"`
	VarName string `mapstructure:"var_name,omitempty"`
	Type    string `mapstructure:"type,omitempty"`
}

func (m MinComponent) APIType() models.AppComponentType {
	switch m.Type {
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

func (c Component) toMinComponent() (MinComponent, error) {
	var minComponent MinComponent
	if err := mapstructure.Decode(c, &minComponent); err != nil {
		return MinComponent{}, err
	}

	return minComponent, nil
}

func (c Component) Name() string {
	minComponent, err := c.toMinComponent()
	if err != nil {
		return ""
	}

	return minComponent.Name
}

func (c Component) AddDependency(val string) {
	var deps []string
	obj, ok := c["dependencies"]
	if !ok {
		deps = make([]string, 0)
	} else {
		obj, ok = obj.([]string)
		if !ok {
			deps = make([]string, 0)
		}
	}

	c["dependencies"] = append(deps, val)
}

func (c Component) ToResourceType() string {
	minComponent, err := c.toMinComponent()
	if err != nil {
		return ""
	}

	switch minComponent.Type {
	case TerraformModuleComponentType:
		return "nuon_terraform_module_component"
	case HelmChartComponentType:
		return "nuon_helm_chart_component"
	case DockerBuildComponentType:
		return "nuon_docker_build_component"
	case ContainerImageComponentType:
		return "nuon_container_image_component"
	case JobComponentType:
		return "nuon_job_component"
	default:
		return ""
	}

	return ""
}

type genericComponent interface {
	ToResource() (map[string]interface{}, error)
	parse(ConfigContext) error
}

func (c Component) ToResource() (map[string]interface{}, error) {
	minComponent, err := c.toMinComponent()
	if err != nil {
		return nil, err
	}

	var (
		cfg  map[string]interface{}
		comp genericComponent
	)

	// grab the actual fields from the components
	switch minComponent.Type {
	case TerraformModuleComponentType:
		var obj TerraformModuleComponentConfig
		if err = mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse terraform module: %w", err)
		}
		comp = &obj
	case HelmChartComponentType:
		var obj HelmChartComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse helm chart: %w", err)
		}
		comp = &obj
	case DockerBuildComponentType:
		var obj DockerBuildComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse docker build: %w", err)
		}
		comp = &obj
	case ContainerImageComponentType:
		var obj ExternalImageComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse external image: %w", err)
		}
		comp = &obj
	case JobComponentType:
		var obj JobComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse job component: %w", err)
		}
		comp = &obj
	default:
		return nil, &ErrConfig{
			Description: "invalid type, must be one of (job, container_image, docker_build, terraform_module, helm_chart)",
			Err:         fmt.Errorf("invalid component type: %s", c["type"]),
		}
	}
	if err := comp.parse(ConfigContextSource); err != nil {
		return nil, fmt.Errorf("unable to parse: %w", err)
	}

	cfg, err = comp.ToResource()
	if err != nil {
		return nil, fmt.Errorf("unable to convert to terraform resource: %w", err)
	}

	return nestWithName(minComponent.Name, cfg), nil
}

func (c Component) parse(ctx ConfigContext) error {
	if ctx != ConfigContextSource {
		return nil
	}

	minComponent, err := c.toMinComponent()
	if err != nil {
		return err
	}
	if minComponent.Source == "" {
		return err
	}

	obj, err := source.LoadSource(minComponent.Source)
	if err != nil {
		return fmt.Errorf("unable to load source: %w", err)
	}
	for k, v := range obj {
		c[k] = v
	}

	return nil
}
