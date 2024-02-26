package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// Component is a flattened configuration type that allows us to define components using a `type: type` field.
type Component map[string]interface{}

type MinComponent struct {
	Name string `mapstructure:"name"`
	Type string `mapstructure:"type"`
}

func (c Component) Name() string {
	var minComponent MinComponent
	if err := mapstructure.Decode(c, &minComponent); err != nil {
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
	var minComponent MinComponent
	if err := mapstructure.Decode(c, &minComponent); err != nil {
		return ""
	}

	switch minComponent.Type {
	case "terraform_module":
		return "nuon_terraform_module_component"
	case "helm_chart":
		return "nuon_helm_chart_component"
	case "docker_build":
		return "nuon_docker_build_component"
	case "container_image":
		return "nuon_container_image_component"
	case "job":
		return "nuon_job_component"
	default:
		return ""
	}

	return ""
}

func (c Component) ToResource() (map[string]interface{}, error) {
	var minComponent MinComponent
	if err := mapstructure.Decode(c, &minComponent); err != nil {
		return nil, fmt.Errorf("invalid component: %w", err)
	}

	var (
		cfg map[string]interface{}
		err error
	)

	// grab the actual fields from the components
	switch minComponent.Type {
	case "terraform_module":
		var obj TerraformModuleComponentConfig
		if err = mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse terraform module: %w", err)
		}
		cfg, err = obj.ToResource()
	case "helm_chart":
		var obj HelmChartComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse helm chart: %w", err)
		}
		cfg, err = obj.ToResource()
	case "docker_build":
		var obj DockerBuildComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse docker build: %w", err)
		}
		cfg, err = obj.ToResource()
	case "container_image":
		var obj ExternalImageComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse external image: %w", err)
		}
		cfg, err = obj.ToResource()
	case "job":
		var obj JobComponentConfig
		if err := mapstructure.Decode(c, &obj); err != nil {
			return nil, fmt.Errorf("unable to parse job component: %w", err)
		}
		cfg, err = obj.ToResource()
	default:
		return nil, &stderr.ErrUser{
			Description: "invalid type, must be one of (job, container_image, docker_build, terraform_module, helm_chart)",
			Err:         fmt.Errorf("invalid component type: %s", c["type"]),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to convert object to map structure: %w", err)
	}

	return nestWithName(minComponent.Name, cfg), nil
}
