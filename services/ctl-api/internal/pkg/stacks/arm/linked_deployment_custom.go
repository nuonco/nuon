package arm

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

var armLogicalIDRegexp = regexp.MustCompile(`[^A-Za-z0-9]`)

func sanitizeDeploymentName(name string) string {
	camel := strcase.ToCamel(name)
	return armLogicalIDRegexp.ReplaceAllString(camel, "")
}

// customDeploymentIdentity holds info about a managed identity declared in a
// custom nested stack so the parent template can create the subscription-level
// role assignment that the identity needs.
type customDeploymentIdentity struct {
	DeploymentName string
	// SanitizedName is the stack's own name without the install namespace
	// DeploymentName may carry. Role naming keys off this, so that it can be
	// namespaced identically at both scopes — see getCustomDeploymentRoleAssignment.
	SanitizedName string
	// output holding the identity's principalId, resolved from the template
	PrincipalIDOutput string
}

// resolvePrincipalIDOutput finds the output carrying a managed identity's
// principalId. An exact "identityPrincipalId" wins; otherwise any output with
// that suffix matches, so a template may prefix it to keep outputs unique.
func resolvePrincipalIDOutput(outputKeys []string) (string, bool) {
	const want = "identityprincipalid"

	for _, key := range outputKeys {
		if strings.EqualFold(key, want) {
			return key, true
		}
	}
	for _, key := range outputKeys {
		if strings.HasSuffix(strings.ToLower(key), want) {
			return key, true
		}
	}

	return "", false
}

// customDeploymentOutputs records a custom stack's deployment name and its
// template's output keys so the phone-home script can report them.
type customDeploymentOutputs struct {
	StackName      string
	DeploymentName string
	OutputKeys     []string
}

func (t *Templates) getCustomLinkedDeployments(inp *stacks.TemplateInput) ([]any, map[string]ARMParameter, []customDeploymentIdentity, []customDeploymentOutputs, error) {
	stacks := inp.AppCfg.StackConfig.CustomNestedStacks
	if len(stacks) == 0 {
		return nil, nil, nil, nil, nil
	}

	// Sort by index
	sorted := make([]config.CustomNestedStack, len(stacks))
	copy(sorted, stacks)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Index < sorted[j].Index
	})

	// Check for duplicate indices
	seenIndices := map[int]string{}
	for _, stack := range sorted {
		if prev, exists := seenIndices[stack.Index]; exists {
			return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks: duplicate index %d for stacks %q and %q", stack.Index, prev, stack.Name)
		}
		seenIndices[stack.Index] = stack.Name
	}

	var resources []any
	var identities []customDeploymentIdentity
	var outputsMeta []customDeploymentOutputs
	hoistedParams := map[string]ARMParameter{}
	allParamNames := map[string]string{}
	prevDeploymentName := ""

	scope := scopeFor(inp)
	vnetDeployment := scope.vnetDeploymentName(inp.Install.ID)

	// param name -> producing deployment; seeded with vnet outputs so they win over custom ones
	wiredOutputs := map[string]string{}
	for _, name := range vnetContractOutputs {
		wiredOutputs[name] = vnetDeployment
	}

	for i, stack := range sorted {
		if stack.Name == "" {
			return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		if stack.TemplateURL == "" {
			return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name)
		}

		sanitizedName := sanitizeDeploymentName(stack.Name)
		deploymentName := scope.customStackDeploymentName(inp.Install.ID, sanitizedName)
		if deploymentName == "" {
			return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): name produces invalid deployment name", i, stack.Name)
		}
		// Caught here rather than at deploy time, where ARM reports an over-long
		// name as a validation failure on a resource the customer never wrote.
		if over := len(deploymentName) - maxARMDeploymentNameLen; over > 0 {
			return nil, nil, nil, nil, fmt.Errorf(
				"custom_nested_stacks[%d] (%s): deployment name %q is %d characters, %d over ARM's limit of %d; shorten the stack name by %d characters",
				i, stack.Name, deploymentName, len(deploymentName), over, maxARMDeploymentNameLen, over,
			)
		}

		// Resolve template URL (use uploaded S3 URL if contents were uploaded)
		templateURL := stack.TemplateURL
		if stack.ContentsHash != "" && t.cfg.AWSCloudFormationStackTemplateBaseURL != "" {
			templateURL = cloudformation.CustomNestedStackTemplateURL(
				t.cfg.AWSCloudFormationStackTemplateBaseURL,
				inp.Install.OrgID,
				inp.AppCfg.AppID,
				stack.ContentsHash,
				stack.TemplateURL,
			)
		}

		// Fetch template, validate structure, and extract parameters
		armTmpl, err := fetchARMTemplate(templateURL)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): %w", i, stack.Name, err)
		}

		if err := validateARMTemplate(armTmpl); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): %w", i, stack.Name, err)
		}

		params, defaultParams := extractARMParameters(armTmpl, ReservedParamNames)

		outputKeys := make([]string, 0, len(armTmpl.Outputs))
		for key := range armTmpl.Outputs {
			outputKeys = append(outputKeys, key)
		}
		sort.Strings(outputKeys)
		outputsMeta = append(outputsMeta, customDeploymentOutputs{
			StackName:      stack.Name,
			DeploymentName: deploymentName,
			OutputKeys:     outputKeys,
		})

		// Track custom nested stacks that declare managed identities so the
		// parent template can create subscription-level role assignments.
		if armTmpl.hasManagedIdentity() {
			principalIDOutput, ok := resolvePrincipalIDOutput(outputKeys)
			if !ok {
				// Caught here rather than at deploy time, where ARM reports it
				// as a missing output on a Nuon-generated resource.
				return nil, nil, nil, nil, fmt.Errorf(
					"custom_nested_stacks[%d] (%s): declares a managed identity but no output named %q (found: %v); add one so the subscription-level role assignment can read its principalId",
					i, stack.Name, "identityPrincipalId", outputKeys,
				)
			}

			// The identity's role deployment carries a longer name than the stack's
			// own, and is install-namespaced at both scopes, so it can exceed the
			// limit even where deploymentName does not.
			roleDeployment := customStackRoleDeploymentName(inp.Install.ID, sanitizedName)
			if over := len(roleDeployment) - maxARMDeploymentNameLen; over > 0 {
				return nil, nil, nil, nil, fmt.Errorf(
					"custom_nested_stacks[%d] (%s): declares a managed identity, whose role deployment name %q is %d characters, %d over ARM's limit of %d; shorten the stack name by %d characters",
					i, stack.Name, roleDeployment, len(roleDeployment), over, maxARMDeploymentNameLen, over,
				)
			}

			identities = append(identities, customDeploymentIdentity{
				DeploymentName:    deploymentName,
				SanitizedName:     sanitizedName,
				PrincipalIDOutput: principalIDOutput,
			})
		}

		// Build deployment parameters
		deploymentParams := map[string]any{}

		// Inject Nuon-reserved params if template declares them
		nuonParams := map[string]string{
			"nuonInstallID": inp.Install.ID,
			"nuonOrgID":     inp.Runner.OrgID,
			"nuonAppID":     inp.Install.AppID,
			// Evaluated in the root, so it has to follow the root's declaration of the
			// region — a parameter at resource-group scope, a variable at subscription
			// scope where it is hidden from the portal's deployment form.
			"location": scopeFor(inp).rootLocationRef(),
		}
		for paramName := range params {
			if val, ok := nuonParams[paramName]; ok {
				deploymentParams[paramName] = map[string]any{"value": val}
			}
		}

		// Explicit parameter values. These arrive already rendered -- see
		// config.RenderCustomNestedStackParameters, called when the install stack
		// version is generated -- so they are used verbatim. The install-input
		// reference form is still resolved here as a fallback for callers that read
		// the config without rendering it first.
		for cfnParamName, templateValue := range stack.Parameters {
			resolved := templateValue
			if inputName, err := config.ParseInstallInputReference(templateValue); err == nil {
				resolved = ""
				if inp.Install.CurrentInstallInputs != nil {
					if val, ok := inp.Install.CurrentInstallInputs.Values[inputName]; ok && val != nil {
						resolved = *val
					}
				}
			}

			deploymentParams[cfnParamName] = map[string]any{"value": resolved}
			delete(defaultParams, cfnParamName)
		}

		// Wire VNet and earlier custom stack outputs to matching parameter names
		for paramName := range params {
			if _, alreadySet := deploymentParams[paramName]; alreadySet {
				continue
			}
			sourceDeployment, ok := wiredOutputs[paramName]
			if !ok {
				continue
			}

			deploymentParams[paramName] = map[string]any{
				"value": fmt.Sprintf("[reference('%s').outputs.%s.value]", sourceDeployment, paramName),
			}
			delete(defaultParams, paramName)
		}

		// Runs after pruning: only names that actually reach the parent can conflict
		for paramName := range defaultParams {
			if owner, exists := allParamNames[paramName]; exists {
				return nil, nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q conflicts with stack %q", i, stack.Name, paramName, owner)
			}
			allParamNames[paramName] = stack.Name
		}

		// An expression default is computed in the template's own parameter scope;
		// hoisting it re-binds it to the parent's. See hoistableDefault.
		for paramName, param := range defaultParams {
			if !hoistableDefault(param.DefaultValue) {
				delete(defaultParams, paramName)
				delete(allParamNames, paramName)
			}
		}

		// Remaining params are hoisted into parent
		for paramName := range defaultParams {
			if _, alreadySet := deploymentParams[paramName]; !alreadySet {
				deploymentParams[paramName] = map[string]any{"value": fmt.Sprintf("[parameters('%s')]", paramName)}
			}
		}

		// Merge hoisted params
		for k, v := range defaultParams {
			hoistedParams[k] = v
		}

		// Sequential: output wiring needs every earlier stack to be a transitive dependency
		var dependsOn []string
		if prevDeploymentName != "" {
			dependsOn = append(dependsOn, prevDeploymentName)
		} else {
			dependsOn = append(dependsOn, vnetDeployment)
			if !t.cfg.UseLocalRunners {
				dependsOn = append(dependsOn, "runnerDeployment")
			}
		}

		deployment := map[string]any{
			"type":       "Microsoft.Resources/deployments",
			"apiVersion": "2022-09-01",
			"name":       deploymentName,
			"dependsOn":  dependsOn,
			"properties": map[string]any{
				"mode": "Incremental",
				"templateLink": map[string]any{
					"uri": templateURL,
				},
				"parameters": deploymentParams,
			},
		}

		// Guessing a stack's scope wrong fails the whole deployment with
		// InvalidScope, so take it from the template's own $schema, as
		// getVNetLinkedDeployment does.
		if isSubscriptionScopedTemplate(armTmpl) {
			scope.targetSubscription(deployment)
		} else {
			scope.targetInstallRG(deployment)
		}

		resources = append(resources, deployment)

		// Registered after the deployment so a stack cannot wire to its own outputs
		for _, key := range outputKeys {
			// a custom output must not shadow a Nuon-injected param for later stacks
			if slices.Contains(ReservedParamNames, key) {
				continue
			}
			if _, exists := wiredOutputs[key]; !exists {
				wiredOutputs[key] = deploymentName
			}
		}

		prevDeploymentName = deploymentName
	}

	return resources, hoistedParams, identities, outputsMeta, nil
}
