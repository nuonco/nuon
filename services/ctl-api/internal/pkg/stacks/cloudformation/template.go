package cloudformation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) Template(inputs *stacks.TemplateInput) (*cloudformation.Template, string, error) {
	if !inputs.CustomStacksOnly {
		if err := config.ValidateCloudFormationCustomNestedStacks("aws-cloudformation", inputs.AppCfg.StackConfig.CustomNestedStacks); err != nil {
			return nil, "", err
		}
	}
	if !inputs.CustomStacksOnly && !inputs.BundleCompile && inputs.Install.InstallConfig != nil {
		known := map[string]bool{"VPC": true, "RunnerAutoScalingGroup": !t.cfg.UseLocalRunners}
		for _, stack := range inputs.AppCfg.StackConfig.CustomNestedStacks {
			known[stack.Name] = true
		}
		for name := range inputs.Install.InstallConfig.AWSNestedStackParameters {
			if inputs.Install.CloudConnectionID == nil {
				return nil, "", fmt.Errorf("nested stack parameter overrides require an AWS cloud connection")
			}
			if !known[name] {
				return nil, "", fmt.Errorf("cloudformation: nested stack %q does not exist", name)
			}
		}
	}

	tmpl, err := t.getAWSTemplate(inputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to create aws template")
	}

	// Marshal the template to JSON
	jsonBytes, err := json.Marshal(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal template to JSON")
	}

	hash := sha256.Sum256(jsonBytes)
	checksum := hex.EncodeToString(hash[:])

	return tmpl, checksum, nil
}

func (t *Templates) applyNestedStackParameters(inp *stacks.TemplateInput, name, templateURL string, parameters map[string]string, defaults map[string]cloudformation.Parameter) error {
	if inp.CustomStacksOnly || inp.BundleCompile || inp.Install.InstallConfig == nil {
		return nil
	}
	overrides := maps.Clone(inp.Install.InstallConfig.AWSNestedStackParameters[name].Parameters)
	if len(overrides) == 0 {
		return nil
	}
	tmpl, err := t.fetchTemplate(templateURL)
	if err != nil {
		return fmt.Errorf("nested stack %q: %w", name, err)
	}
	for parameter := range overrides {
		if _, ok := tmpl.Parameters[parameter]; !ok {
			return fmt.Errorf("nested stack %q: parameter %q is not declared in the template", name, parameter)
		}
	}
	stateData, err := inp.InstallState.AsMap()
	if err != nil {
		return errors.Wrap(err, "unable to get install state for nested stack parameters")
	}
	if err := render.RenderTextStringMap(overrides, stateData); err != nil {
		return fmt.Errorf("unable to render nested stack %q parameters: %w", name, err)
	}
	for key, value := range overrides {
		parameters[key] = value
		delete(defaults, key)
	}
	return nil
}
