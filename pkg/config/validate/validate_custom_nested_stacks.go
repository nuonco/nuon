package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/pkg/config"
)

type cfnOutputsShape struct {
	Outputs map[string]struct{} `yaml:"Outputs" json:"Outputs"`
}

func fetchTemplateOutputs(templateURL string) (map[string]struct{}, error) {
	resp, err := http.Get(templateURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, templateURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tmpl cfnOutputsShape
	if strings.HasSuffix(templateURL, ".json") {
		err = json.Unmarshal(body, &tmpl)
	} else {
		err = yaml.Unmarshal(body, &tmpl)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Outputs, nil
}

func ValidateCustomNestedStackOutputs(a *config.AppConfig) error {
	if a.Stack == nil || len(a.Stack.CustomNestedStacks) == 0 {
		return nil
	}

	sorted := make([]config.CustomNestedStack, len(a.Stack.CustomNestedStacks))
	copy(sorted, a.Stack.CustomNestedStacks)
	slices.SortStableFunc(sorted, func(a, b config.CustomNestedStack) int {
		return a.Index - b.Index
	})

	fcOutputs := map[string]string{}
	for resourceName, templateURL := range map[string]string{
		"VPC":                    a.Stack.VPCNestedTemplateURL,
		"RunnerAutoScalingGroup": a.Stack.RunnerNestedTemplateURL,
	} {
		if templateURL == "" {
			continue
		}
		outputs, err := fetchTemplateOutputs(templateURL)
		if err != nil {
			continue
		}
		for outputName := range outputs {
			fcOutputs[outputName] = resourceName
		}
	}

	additionalOutputs := map[string]string{}

	var warnings []string
	for _, stack := range sorted {
		if stack.TemplateURL == "" {
			continue
		}

		outputs, err := fetchTemplateOutputs(stack.TemplateURL)
		if err != nil {
			continue
		}

		for outputName := range outputs {
			if fcResource, exists := fcOutputs[outputName]; exists {
				warnings = append(warnings, fmt.Sprintf(
					"custom_nested_stacks (%s): output %q collides with %s stack output (first-class output will take precedence)",
					stack.Name, outputName, fcResource,
				))
			} else if prevStack, exists := additionalOutputs[outputName]; exists {
				warnings = append(warnings, fmt.Sprintf(
					"custom_nested_stacks (%s): output %q also declared by stack %q (earlier stack output will be used)",
					stack.Name, outputName, prevStack,
				))
			}

			if _, isFirstClass := fcOutputs[outputName]; !isFirstClass {
				if _, seen := additionalOutputs[outputName]; !seen {
					additionalOutputs[outputName] = stack.Name
				}
			}
		}
	}

	if len(warnings) == 0 {
		return nil
	}

	return config.ErrConfig{
		Description: strings.Join(warnings, "; "),
		Warning:     true,
		Err:         fmt.Errorf("custom nested stack output warnings: %s", strings.Join(warnings, "; ")),
	}
}
