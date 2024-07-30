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

func (c Component) Parse() (interface{}, string, error) {
	minComponent, err := c.toMinComponent()
	if err != nil {
		return nil, "", err
	}

	switch minComponent.Type {
	case ContainerImageComponentType:
		var containerImage ExternalImageComponentConfig
		if err := mapstructure.Decode(c, &containerImage); err != nil {
			return nil, "", ErrConfig{
				Description: fmt.Sprintf("unable to parse container image: %s", err.Error()),
			}
		}
	case DockerBuildComponentType:
		var obj DockerBuildComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, "", ErrConfig{
				Description: fmt.Sprintf("unable to parse docker build: %s", err.Error()),
			}
		}
	case HelmChartComponentType:
		var obj HelmChartComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, "", ErrConfig{
				Description: fmt.Sprintf("unable to parse helm chart: %s", err.Error()),
			}
		}
	case TerraformModuleComponentType:
		var obj TerraformModuleComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, "", ErrConfig{
				Description: fmt.Sprintf("unable to parse terraform module component: %s", err.Error()),
			}
		}
	case JobComponentType:
		var obj JobComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, "", ErrConfig{
				Description: fmt.Sprintf("unable to parse job component: %s", err.Error()),
			}
		}
	}

	return nil, "", ErrConfig{Description: "invalid type"}
}

type MinComponent struct {
	Source  string `mapstructure:"source,omitempty"`
	Name    string `mapstructure:"name" jsonschema:"required"`
	VarName string `mapstructure:"var_name,omitempty"`
	Type    string `mapstructure:"type,omitempty" jsonschema:"required"`
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

type genericComponent interface {
	parse(ConfigContext) error
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
		return nil
	}

	obj, err := source.LoadSource(minComponent.Source)
	if err != nil {
		return ErrConfig{
			Description: "unable to load source",
			Err:         fmt.Errorf("unable to load source: %w", err),
		}
	}
	for k, v := range obj {
		c[k] = v
	}

	return nil
}
