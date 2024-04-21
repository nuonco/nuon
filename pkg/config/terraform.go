package config

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	DefaultTerraformVersion string = "1.7.5"
	DefaultModuleFileName   string = "module.tf.json"
)

func (a *AppConfig) ToTerraformJSON(backendType BackendType) ([]byte, error) {
	json := newJSON()
	tfMap, err := a.ToTerraform(backendType)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to terraform: %w", err)
	}

	byts, err := json.Marshal(tfMap)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal config to json: %w", err)
	}

	return byts, nil
}

func (a *AppConfig) ToTerraform(backendType BackendType) (map[string]interface{}, error) {
	resources := []resource{
		a.Sandbox,
		a.Runner,
	}

	if a.Inputs != nil && len(a.Inputs.Inputs) > 0 {
		resources = append(resources, a.Inputs)
	}
	if a.Installer != nil && *a.Installer != (AppInstallerConfig{}) {
		resources = append(resources, a.Installer)
	}

	for idx, comp := range a.Components {
		if idx > 0 {
			prevComp := a.Components[idx-1]
			prevCompID := fmt.Sprintf("${%s.%s.id}", prevComp.ToResourceType(), prevComp.Name())
			comp.AddDependency(prevCompID)
		}

		resources = append(resources, comp)
	}

	tfResources := map[string]interface{}{}
	for _, resource := range resources {
		if resource == nil {
			continue
		}

		typ := resource.ToResourceType()

		tfResource, err := resource.ToResource()
		if err != nil {
			return nil, fmt.Errorf("unable to convert %s to terraform resource: %w", typ, err)
		}

		_, exists := tfResources[typ]
		if exists {
			tfResources[typ] = generics.MergeMap(tfResources[typ].(map[string]interface{}), tfResource)
		} else {
			tfResources[typ] = tfResource
		}
	}

	backend := "s3"
	if backendType == BackendTypeLocal {
		backend = "local"
	}

	return map[string]interface{}{
		"terraform": map[string]interface{}{
			"required_version": ">= 1.5.3",
			"backend": map[string]interface{}{
				backend: map[string]interface{}{},
			},
			"required_providers": map[string]interface{}{
				"nuon": map[string]interface{}{
					"source":  "nuonco/nuon",
					"version": ">= 0.12.0",
				},
			},
		},
		"provider": map[string]interface{}{
			"nuon": map[string]interface{}{},
		},
		"variable": map[string]interface{}{
			"app_id": map[string]interface{}{
				"type": "string",
			},
		},
		"resource": tfResources,
	}, nil
}
