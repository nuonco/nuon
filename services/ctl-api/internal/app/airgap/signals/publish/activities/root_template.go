package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"sigs.k8s.io/yaml"

	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// The bundled root template is rewritten to use customer-hosted nested templates during stack preparation.
func (a *Activities) rootTemplateInputs(ctx context.Context, orgID, installID, syntheticInstallID string, cfg *app.AppConfig) ([]byte, string, error) {
	if installID == "" {
		return a.compileRootTemplate(ctx, cfg, syntheticInstallID)
	}
	var version app.InstallStackVersion
	err := a.db.WithContext(ctx).
		Where(app.InstallStackVersion{OrgID: orgID, InstallID: installID}).
		Order("created_at DESC").
		First(&version).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", fmt.Errorf("load install stack version for reference install %s: %w", installID, err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return a.compileRootTemplate(ctx, cfg, syntheticInstallID)
	}
	if len(version.Contents) == 0 {
		return nil, "", fmt.Errorf("reference install %s stack version %s has no rendered template contents; regenerate its stack version before publishing", installID, version.ID)
	}
	contents, err := prepareRootTemplateForBundle(version.Contents)
	if err != nil {
		return nil, "", fmt.Errorf("prepare root stack template from stack version %s: %w", version.ID, err)
	}
	source := version.TemplateURL
	if source == "" {
		source = "install-stack-version:" + version.ID
	}
	return contents, source, nil
}

const (
	rootTemplateAssetRole = "root"

	stackAssetJSONMediaType = "application/json"
)

// CloudFormation never resolves these, so leaking either one deploys a stack
// with garbage physical names (IAM rejects "{{.nuon.install.id}}-provision")
// or stale placeholder values. Fail the publish instead.
var unrenderedNuonTemplate = regexp.MustCompile(`\{\{[^{}]*\.nuon[^{}]*\}\}`)

// Air-gapped root templates must contain no runner API or phone-home control-plane credentials.
func prepareRootTemplateForBundle(contents []byte) ([]byte, error) {
	if match := unrenderedNuonTemplate.Find(contents); match != nil {
		return nil, fmt.Errorf("install stack template contains an unrendered template expression %q; the app config references install state that is unavailable when compiling a stack template for an air-gapped bundle", match)
	}
	if bytes.Contains(contents, []byte(runnerairgap.InputPlaceholderPrefix)) {
		return nil, fmt.Errorf("install stack template references install inputs, which are only late-bound inside plans; remove install input references from the stack, permissions, break-glass, and secrets config")
	}
	var doc map[string]any
	if err := json.Unmarshal(contents, &doc); err != nil {
		return nil, fmt.Errorf("decode install stack template: %w", err)
	}
	params, _ := doc["Parameters"].(map[string]any)
	if _, ok := params["PhoneHomeS3Bucket"]; !ok {
		return nil, fmt.Errorf("install stack template has no PhoneHomeS3Bucket parameter; regenerate the reference install's stack version with an S3-capable control plane before publishing")
	}
	if err := validatePhoneHomeScriptSupportsS3(doc); err != nil {
		return nil, err
	}
	removeRunnerAPISurfaces(doc)
	removePhoneHomeControlPlaneSurfaces(doc)
	return json.Marshal(doc)
}

// A template parameter alone is insufficient because older embedded scripts ignore the S3 rendezvous.
func validatePhoneHomeScriptSupportsS3(doc map[string]any) error {
	resources, _ := doc["Resources"].(map[string]any)
	lambda, _ := resources["RunnerPhoneHome"].(map[string]any)
	properties, _ := lambda["Properties"].(map[string]any)
	code, _ := properties["Code"].(map[string]any)
	script, _ := code["ZipFile"].(string)
	if script == "" {
		return fmt.Errorf("install stack template has no inline RunnerPhoneHome Lambda source")
	}
	if !strings.Contains(script, "NUON_PHONE_HOME_S3_BUCKET") {
		return fmt.Errorf("install stack template's phone-home script does not support the S3 rendezvous; regenerate the reference install's stack version with an S3-capable phone-home script before publishing")
	}
	return nil
}

// Reject templates requiring RunnerApiToken because the exported root template intentionally omits it.
func validateRunnerNestedTemplateAirgapCompatible(data []byte) error {
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
		return fmt.Errorf("runner nested template requires a RunnerApiToken parameter, but air-gapped bundles carry no runner API token; use an airgap-compatible runner nested template (v0.4.0+) that does not require it")
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

// Blanking the fallback URL makes missing S3 parameters fail closed instead of contacting the vendor control plane.
func removePhoneHomeControlPlaneSurfaces(doc map[string]any) {
	resources, _ := doc["Resources"].(map[string]any)

	props, _ := resources["PhoneHomeProps"].(map[string]any)
	if properties, ok := props["Properties"].(map[string]any); ok {
		if _, ok := properties["url"]; ok {
			properties["url"] = ""
		}
	}

	lambda, _ := resources["RunnerPhoneHome"].(map[string]any)
	lambdaProperties, _ := lambda["Properties"].(map[string]any)
	environment, _ := lambdaProperties["Environment"].(map[string]any)
	if variables, ok := environment["Variables"].(map[string]any); ok {
		delete(variables, "NUON_PHONE_HOME_SECRET_ARN")
		delete(variables, "NUON_PHONE_HOME_SECRET_REGION")
	}

	role, _ := resources["RunnerPhoneHomeRole"].(map[string]any)
	roleProperties, _ := role["Properties"].(map[string]any)
	if policies, ok := roleProperties["Policies"].([]any); ok {
		filtered := policies[:0]
		for _, item := range policies {
			if policy, ok := item.(map[string]any); ok {
				if name, ok := policy["PolicyName"].(string); ok && name == "PhoneHomeSecretPolicy" {
					continue
				}
			}
			filtered = append(filtered, item)
		}
		roleProperties["Policies"] = filtered
	}
}
