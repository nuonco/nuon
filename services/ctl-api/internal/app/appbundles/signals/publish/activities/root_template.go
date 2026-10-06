package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"sigs.k8s.io/yaml"

	appbundle "github.com/nuonco/nuon/pkg/appbundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appbundles "github.com/nuonco/nuon/services/ctl-api/internal/app/appbundles"
)

// The bundled root template is always compiled from the app config against a
// synthetic install; bundle publishes never reuse a rendered install template.
func (a *Activities) rootTemplateInputs(ctx context.Context, syntheticInstallID string, cfg *app.AppConfig, runnerImageTag string) ([]byte, string, []appbundles.Finding, error) {
	return a.compileRootTemplate(ctx, cfg, syntheticInstallID, runnerImageTag)
}

const (
	rootTemplateAssetRole = "root"

	stackAssetJSONMediaType = "application/json"
)

// CloudFormation never resolves these, so leaking either one deploys a stack
// with garbage physical names (IAM rejects "{{.nuon.install.id}}-provision")
// or stale placeholder values. Fail the publish instead.
var unrenderedNuonTemplate = regexp.MustCompile(`\{\{[^{}]*\.nuon[^{}]*\}\}`)

// Runner API surfaces are stripped because the compiled values are synthetic
// ("compiled" tokens); phone-home stays
// pointed at the vendor control plane, since bundle installs may stay
// connected and the S3 rendezvous is optional.
func prepareRootTemplateForBundle(contents []byte) ([]byte, error) {
	if match := unrenderedNuonTemplate.Find(contents); match != nil {
		return nil, fmt.Errorf("install stack template contains an unrendered template expression %q; the app config references install state that is unavailable when compiling a stack template for a portable bundle", match)
	}
	if bytes.Contains(contents, []byte(appbundle.InputPlaceholderPrefix)) {
		return nil, fmt.Errorf("install stack template references install inputs, which are only late-bound inside plans; remove install input references from the stack, permissions, break-glass, and secrets config")
	}
	var doc map[string]any
	if err := json.Unmarshal(contents, &doc); err != nil {
		return nil, fmt.Errorf("decode install stack template: %w", err)
	}
	removeRunnerAPISurfaces(doc)
	return json.Marshal(doc)
}

// Reject templates requiring RunnerApiToken because the exported root template intentionally omits it.
func validateRunnerNestedTemplateOfflineCompatible(data []byte) error {
	var doc struct {
		Parameters map[string]map[string]any `json:"Parameters"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("decode runner nested template: %w", err)
	}
	spec, ok := doc.Parameters["RunnerApiToken"]
	if !ok {
		return nil
	}
	if _, hasDefault := spec["Default"]; !hasDefault {
		return fmt.Errorf("runner nested template requires a RunnerApiToken parameter, but offline customer-managed installs carry no runner API token; use an offline-compatible runner nested template (v0.4.0+) that does not require it")
	}
	return nil
}

func removeRunnerAPISurfaces(value any) {
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			if _, isString := child.(string); isString && (key == "RunnerApiToken" || key == "RunnerApiUrl") {
				delete(node, key)
				continue
			}
			if list, isList := child.([]any); isList {
				node[key] = withoutRunnerAPITags(list)
				removeRunnerAPISurfaces(node[key])
				continue
			}
			removeRunnerAPISurfaces(child)
		}
	case []any:
		for _, child := range node {
			removeRunnerAPISurfaces(child)
		}
	}
}

func withoutRunnerAPITags(list []any) []any {
	filtered := list[:0]
	for _, item := range list {
		if entry, ok := item.(map[string]any); ok {
			if key, ok := entry["Key"].(string); ok && (key == "nuon_runner_api_token" || key == "nuon_runner_api_url") {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered
}
