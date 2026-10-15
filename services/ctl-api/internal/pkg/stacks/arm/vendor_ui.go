package arm

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) loadStackUIDefinition(inp *stacks.TemplateInput) ([]byte, error) {
	if inp == nil || inp.AppCfg == nil || inp.AppCfg.StackConfig.UIDefinition == nil {
		return nil, nil
	}
	ui := inp.AppCfg.StackConfig.UIDefinition
	if strings.TrimSpace(ui.Contents) != "" {
		return []byte(ui.Contents), nil
	}
	if ui.SourceURL == "" {
		return nil, nil
	}
	return fetchLimitedBody(ui.SourceURL)
}

// vendorUINames is how a vendor createUiDefinition refers to template parameters.
// bound names are controls the portal should collect. hidden names are controls
// with static visible:false; those parameters keep the nested template's default.
type vendorUINames struct {
	bound  map[string]bool
	hidden map[string]bool
}

func (t *Templates) vendorUINames(inp *stacks.TemplateInput) (vendorUINames, error) {
	names := vendorUINames{bound: map[string]bool{}, hidden: map[string]bool{}}
	body, err := t.loadStackUIDefinition(inp)
	if err != nil || len(body) == 0 {
		return names, err
	}

	var doc struct {
		Parameters struct {
			Basics []any `json:"basics"`
			Steps  []any `json:"steps"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return names, fmt.Errorf("unable to parse stack ui definition: %w", err)
	}
	collectNamedElements(doc.Parameters.Basics, names.bound, names.hidden, false)
	collectNamedElements(doc.Parameters.Steps, names.bound, names.hidden, false)
	return names, nil
}

func collectNamedElements(nodes []any, bound, hidden map[string]bool, forceHidden bool) {
	for _, node := range nodes {
		el, ok := node.(map[string]any)
		if !ok {
			continue
		}
		hiddenEl := forceHidden || staticallyHidden(el)
		if children, ok := el["elements"].([]any); ok {
			collectNamedElements(children, bound, hidden, hiddenEl)
			continue
		}
		name, _ := el["name"].(string)
		if name == "" || displayOnlyUIElement(el) {
			continue
		}
		if hiddenEl {
			hidden[name] = true
			continue
		}
		bound[name] = true
	}
}

func staticallyHidden(el map[string]any) bool {
	hidden, ok := el["visible"].(bool)
	return ok && !hidden
}

func displayOnlyUIElement(el map[string]any) bool {
	switch el["type"] {
	case "Microsoft.Common.InfoBox", "Microsoft.Common.TextBlock":
		return true
	default:
		return false
	}
}

func fetchLimitedBody(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ui definition %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching ui definition %s", resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTemplateBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read ui definition %s: %w", rawURL, err)
	}
	if len(body) > maxTemplateBytes {
		return nil, fmt.Errorf("ui definition %s exceeds %d bytes", rawURL, maxTemplateBytes)
	}
	return body, nil
}

// mergeVendorUI applies a vendor createUiDefinition to the generated portal form.
// Nuon keeps parameters.config, the secrets section, and the roles section.
// A vendor element replaces the generated control for the parameter it names.
// Parameters the file does not bind keep the generated control.
func mergeVendorUI(uiDef map[string]any, vendor []byte, wrapperParams map[string]ARMParameter) error {
	var doc struct {
		Parameters struct {
			Basics  []any          `json:"basics"`
			Steps   []any          `json:"steps"`
			Outputs map[string]any `json:"outputs"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(vendor, &doc); err != nil {
		return fmt.Errorf("unable to parse stack ui definition: %w", err)
	}

	params, _ := uiDef["parameters"].(map[string]any)
	if params == nil {
		return fmt.Errorf("generated ui definition is missing parameters")
	}
	generatedBasics, _ := params["basics"].([]any)
	generatedOutputs, _ := params["outputs"].(map[string]any)
	if generatedOutputs == nil {
		generatedOutputs = map[string]any{}
	}

	owned := nuonOwnedUIElements(generatedBasics)
	refs := map[string]string{}
	maps.Copy(refs, indexBasicsElementRefs(doc.Parameters.Basics))
	maps.Copy(refs, indexStepElementRefs(doc.Parameters.Steps))

	bound := map[string]bool{}
	for name := range refs {
		if owned[name] {
			continue
		}
		if _, declared := wrapperParams[name]; !declared {
			continue
		}
		bound[name] = true
	}

	params["basics"] = append(
		filterVendorControls(doc.Parameters.Basics, owned, wrapperParams),
		filterUIElements(generatedBasics, bound)...,
	)
	params["steps"] = filterVendorControls(doc.Parameters.Steps, owned, wrapperParams)

	outputs := maps.Clone(generatedOutputs)
	for name := range bound {
		if vendorOutput, ok := doc.Parameters.Outputs[name].(string); ok && vendorOutput != "" {
			outputs[name] = vendorOutput
			continue
		}
		ref := refs[name]
		output, ok := outputForPlacedElement(name, wrapperParams[name], ref)
		if !ok {
			continue
		}
		outputs[name] = output
	}
	params["outputs"] = outputs
	return nil
}

func nuonOwnedUIElements(basics []any) map[string]bool {
	owned := map[string]bool{
		"location":        true,
		"deployTimestamp": true,
	}
	for _, node := range basics {
		el, ok := node.(map[string]any)
		if !ok || el["type"] != "Microsoft.Common.Section" {
			continue
		}
		name, _ := el["name"].(string)
		if name != "secrets" && name != "roles" {
			continue
		}
		children, _ := el["elements"].([]any)
		for _, childNode := range children {
			child, ok := childNode.(map[string]any)
			if !ok {
				continue
			}
			if childName, _ := child["name"].(string); childName != "" {
				owned[childName] = true
			}
		}
	}
	return owned
}

func indexBasicsElementRefs(basics []any) map[string]string {
	out := map[string]string{}
	for _, node := range basics {
		el, ok := node.(map[string]any)
		if !ok {
			continue
		}
		if staticallyHidden(el) {
			continue
		}
		name, _ := el["name"].(string)
		if el["type"] == "Microsoft.Common.Section" {
			children, _ := el["elements"].([]any)
			for _, childNode := range children {
				child, ok := childNode.(map[string]any)
				if !ok || staticallyHidden(child) {
					continue
				}
				childName, _ := child["name"].(string)
				if childName != "" {
					out[childName] = fmt.Sprintf("basics('%s').%s", name, childName)
				}
			}
			continue
		}
		if name != "" {
			out[name] = fmt.Sprintf("basics('%s')", name)
		}
	}
	return out
}

func indexStepElementRefs(steps []any) map[string]string {
	out := map[string]string{}
	for _, node := range steps {
		step, ok := node.(map[string]any)
		if !ok {
			continue
		}
		if staticallyHidden(step) {
			continue
		}
		stepName, _ := step["name"].(string)
		elements, _ := step["elements"].([]any)
		for _, elementNode := range elements {
			el, ok := elementNode.(map[string]any)
			if !ok || staticallyHidden(el) {
				continue
			}
			name, _ := el["name"].(string)
			if el["type"] == "Microsoft.Common.Section" {
				children, _ := el["elements"].([]any)
				for _, childNode := range children {
					child, ok := childNode.(map[string]any)
					if !ok || staticallyHidden(child) {
						continue
					}
					childName, _ := child["name"].(string)
					if childName != "" {
						out[childName] = fmt.Sprintf("steps('%s').%s.%s", stepName, name, childName)
					}
				}
				continue
			}
			if name != "" {
				out[name] = fmt.Sprintf("steps('%s').%s", stepName, name)
			}
		}
	}
	return out
}

// filterVendorControls keeps display-only elements and controls for parameters
// the wrapper declares. A static visible:false control is removed, and so is a
// text box for a parameter that was not hoisted, so the form does not ask for a
// value the template already computed.
func filterVendorControls(nodes []any, drop map[string]bool, wrapperParams map[string]ARMParameter) []any {
	if len(nodes) == 0 {
		return []any{}
	}
	out := make([]any, 0, len(nodes))
	for _, node := range nodes {
		el, ok := node.(map[string]any)
		if !ok {
			out = append(out, node)
			continue
		}
		if staticallyHidden(el) {
			continue
		}
		if _, hasElements := el["elements"]; hasElements {
			children, _ := el["elements"].([]any)
			filtered := filterVendorControls(children, drop, wrapperParams)
			if len(filtered) == 0 {
				continue
			}
			cloned := maps.Clone(el)
			cloned["elements"] = filtered
			out = append(out, cloned)
			continue
		}
		name, _ := el["name"].(string)
		if drop[name] {
			continue
		}
		if name != "" && !displayOnlyUIElement(el) {
			if _, declared := wrapperParams[name]; !declared {
				continue
			}
		}
		out = append(out, el)
	}
	return out
}

func filterUIElements(nodes []any, drop map[string]bool) []any {
	if len(nodes) == 0 {
		return []any{}
	}
	out := make([]any, 0, len(nodes))
	for _, node := range nodes {
		el, ok := node.(map[string]any)
		if !ok {
			out = append(out, node)
			continue
		}
		if _, hasElements := el["elements"]; hasElements {
			children, _ := el["elements"].([]any)
			filtered := filterUIElements(children, drop)
			if len(filtered) == 0 {
				continue
			}
			cloned := maps.Clone(el)
			cloned["elements"] = filtered
			out = append(out, cloned)
			continue
		}
		name, _ := el["name"].(string)
		if drop[name] {
			continue
		}
		out = append(out, el)
	}
	return out
}

func outputForPlacedElement(name string, p ARMParameter, ref string) (string, bool) {
	_, output, ok := parameterElement(name, p, "", "")
	if !ok {
		return "", false
	}
	return strings.ReplaceAll(output, fmt.Sprintf("basics('%s')", name), ref), true
}
