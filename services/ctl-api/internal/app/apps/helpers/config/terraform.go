package config

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
)

func (a *AppConfig) ToTerraformJSON(env config.Env) ([]byte, error) {
	json := newJSON()
	tfMap, err := a.ToTerraform(env)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to terraform: %w", err)
	}

	byts, err := json.Marshal(tfMap)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal config to json: %w", err)
	}

	return byts, nil
}

func (a *AppConfig) ToTerraform(env config.Env) (map[string]interface{}, error) {
	resources := []resource{
		a.Inputs,
		a.Sandbox,
		a.Runner,
	}
	if a.Installer != nil && *a.Installer != (AppInstallerConfig{}) {
		resources = append(resources, a.Installer)
	}

	//for _, comp := range a.Components {
	////resources = append(resources, comp)
	//}

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

		tfResources[typ] = tfResource
	}

	backendType := "s3"
	if env == config.Development {
		backendType = "local"
	}

	return map[string]interface{}{
		"terraform": map[string]interface{}{
			"required_version": ">= 1.5.3",
			"backend": map[string]interface{}{
				backendType: map[string]interface{}{},
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
