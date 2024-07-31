package config

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

// NOTE(jm): this might not be needed
//func ComponentEncode(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
//if toType != reflect.TypeOf(Component{}) {
//return from, nil
//}
//if toType != reflect.TypeOf(map[string]interface{}{}) {
//return from, nil
//}

// comp := from.(Component)

//var obj map[string]interface{}
//if err := mapstructure.Decode(comp, &obj); err != nil {
//return from, fmt.Errorf("unable to convert object: %w", err)
//}

//return obj, nil
//}

func DecodeComponent(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	if fromType != reflect.TypeOf(map[string]interface{}{}) {
		return from, nil
	}
	if toType != reflect.TypeOf(Component{}) {
		return from, nil
	}

	obj := from.(map[string]interface{})
	src, ok := obj["source"]
	if ok {
		srcObj, err := source.LoadSource(src.(string))
		if err != nil {
			return from, ErrConfig{
				Description: "unable to load source",
				Err:         fmt.Errorf("unable to load source: %w", err),
			}
		}
		for k, v := range srcObj {
			obj[k] = v
		}
	}

	var comp Component
	if err := mapstructure.Decode(obj, &comp); err != nil {
		return from, fmt.Errorf("unable to convert to component: %w", err)
	}
	if comp.Type == ComponentTypeUnknown {
		return from, ErrConfig{
			Description: "must set `type` on each component",
		}
	}

	switch comp.Type {
	case ContainerImageComponentType, ExternalImageComponentType:
		var cmpCfg ExternalImageComponentConfig
		if err := mapstructure.Decode(obj, &cmpCfg); err != nil {
			return from, ErrConfig{
				Description: fmt.Sprintf("unable to parse container image: %s", err.Error()),
			}
		}
		comp.ExternalImage = &cmpCfg
	case DockerBuildComponentType:
		var cmpCfg DockerBuildComponentConfig
		if err := mapstructure.Decode(obj, &cmpCfg); err != nil {
			return from, ErrConfig{
				Description: fmt.Sprintf("unable to parse docker build: %s", err.Error()),
			}
		}
		comp.DockerBuild = &cmpCfg
	case HelmChartComponentType:
		var cmpCfg HelmChartComponentConfig
		if err := mapstructure.Decode(obj, &cmpCfg); err != nil {
			return from, ErrConfig{
				Description: fmt.Sprintf("unable to parse helm chart: %s", err.Error()),
			}
		}
		comp.HelmChart = &cmpCfg
	case TerraformModuleComponentType:
		var cmpCfg TerraformModuleComponentConfig
		if err := mapstructure.Decode(obj, &cmpCfg); err != nil {
			return from, ErrConfig{
				Description: fmt.Sprintf("unable to parse terraform module component: %s", err.Error()),
			}
		}
		comp.TerraformModule = &cmpCfg
	case JobComponentType:
		var cmpCfg JobComponentConfig
		if err := mapstructure.Decode(obj, &cmpCfg); err != nil {
			return from, ErrConfig{
				Description: fmt.Sprintf("unable to parse job component: %s", err.Error()),
			}
		}
		comp.Job = &cmpCfg
	default:
		return from, ErrConfig{Description: "invalid type"}
	}

	return comp, nil
}
