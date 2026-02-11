package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/pkg/types/components/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	policyhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
)

type PrepPolicyEvaluationRequest struct {
	StepTargetID   string `validate:"required"`
	StepTargetType string `validate:"required"`
}

// PolicyViolation is an alias to app.PolicyViolation for use in activity requests/responses.
type PolicyViolation = app.PolicyViolation

type PolicyToEvaluate struct {
	PolicyID      string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	PolicyName    string `json:"policy_name" temporaljson:"policy_name,omitempty"`
	Contents      string `json:"contents" temporaljson:"contents,omitempty"`
	InputJSON     []byte `json:"input_json" temporaljson:"input_json,omitempty"`
	InputIndex    int    `json:"input_index" temporaljson:"input_index,omitempty"`       // Index of the input document
	InputIdentity string `json:"input_identity" temporaljson:"input_identity,omitempty"` // Human-readable input reference
}

type PrepPolicyEvaluationResult struct {
	Policies         []PolicyToEvaluate `json:"policies" temporaljson:"policies,omitempty"`
	HasPolicies      bool               `json:"has_policies" temporaljson:"has_policies,omitempty"`
	OrgID            string             `json:"org_id" temporaljson:"org_id,omitempty"`
	AppID            string             `json:"app_id" temporaljson:"app_id,omitempty"`
	InstallID        *string            `json:"install_id" temporaljson:"install_id,omitempty"`
	InstallSandboxID *string            `json:"install_sandbox_id" temporaljson:"install_sandbox_id,omitempty"`
	ComponentID      *string            `json:"component_id" temporaljson:"component_id,omitempty"`
	ComponentBuildID *string            `json:"component_build_id" temporaljson:"component_build_id,omitempty"`
	PolicyIDs        []string           `json:"policy_ids" temporaljson:"policy_ids,omitempty"`
	InputCount       int                `json:"input_count" temporaljson:"input_count,omitempty"`

	// Human-readable names for display in reports
	OrgName       string `json:"org_name" temporaljson:"org_name,omitempty"`
	AppName       string `json:"app_name" temporaljson:"app_name,omitempty"`
	InstallName   string `json:"install_name" temporaljson:"install_name,omitempty"`
	ComponentName string `json:"component_name" temporaljson:"component_name,omitempty"`
}

// @temporal-gen activity
// @max-retries 1
// @schedule-to-close-timeout 5m
// @start-to-close-timeout 4m
func (a *Activities) PrepPolicyEvaluation(ctx context.Context, req *PrepPolicyEvaluationRequest) (*PrepPolicyEvaluationResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(
		zap.String("step_target_id", req.StepTargetID),
		zap.String("step_target_type", req.StepTargetType),
	)

	l.Info("preparing policy evaluation")

	policyContext, err := a.resolvePolicyContext(ctx, req.StepTargetID, req.StepTargetType)
	if err != nil {
		l.Error("unable to resolve policy context", zap.Error(err))
		return nil, errors.Wrap(err, "unable to resolve policy context")
	}

	l = l.With(
		zap.String("app_config_id", policyContext.AppConfigID),
		zap.String("component_type", string(policyContext.ComponentType)),
		zap.String("component_name", policyContext.ComponentName),
		zap.Bool("is_sandbox", policyContext.IsSandbox),
	)

	policiesConfig, err := a.appsHelpers.GetPoliciesConfigByAppConfigID(ctx, policyContext.AppConfigID)
	if err != nil {
		l.Error("unable to get policies config", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get policies config")
	}

	approvalPlan, err := a.getApprovalPlan(ctx, req.StepTargetID)
	if err != nil {
		l.Error("unable to get plan contents", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get plan contents")
	}

	applicablePolicies := a.filterApplicablePolicies(
		policiesConfig.Policies,
		policyContext.ComponentType,
		policyContext.ComponentName,
		policyContext.IsSandbox,
	)

	l.Info("filtered applicable policies", zap.Int("count", len(applicablePolicies)))

	if len(applicablePolicies) == 0 {
		l.Info("no applicable policies found")
		return &PrepPolicyEvaluationResult{
			Policies:         []PolicyToEvaluate{},
			HasPolicies:      false,
			OrgID:            policyContext.OrgID,
			AppID:            policyContext.AppID,
			InstallID:        policyContext.InstallID,
			InstallSandboxID: policyContext.InstallSandboxID,
			ComponentID:      policyContext.ComponentID,
			ComponentBuildID: policyContext.ComponentBuildID,
			PolicyIDs:        []string{},
			InputCount:       0,
			OrgName:          policyContext.OrgName,
			AppName:          policyContext.AppName,
			InstallName:      policyContext.InstallName,
			ComponentName:    policyContext.ComponentName,
		}, nil
	}

	policyIDs := make([]string, 0, len(applicablePolicies))
	for _, policy := range applicablePolicies {
		policyIDs = append(policyIDs, policy.ID)
	}

	policyInputs, inputIdentities, err := a.preparePolicyInputs(approvalPlan.PlanContents, policyContext)
	if err != nil {
		l.Error("unable to prepare policy inputs", zap.Error(err))
		return nil, errors.Wrap(err, "unable to prepare policy inputs")
	}

	policies := a.buildPolicyEvaluationItems(applicablePolicies, policyInputs, inputIdentities)

	l.Info("policy evaluation preparation complete",
		zap.Int("policies_count", len(applicablePolicies)),
		zap.Int("inputs_count", len(policyInputs)),
		zap.Int("total_evaluations", len(policies)),
	)

	return &PrepPolicyEvaluationResult{
		Policies:         policies,
		HasPolicies:      true,
		OrgID:            policyContext.OrgID,
		AppID:            policyContext.AppID,
		InstallID:        policyContext.InstallID,
		InstallSandboxID: policyContext.InstallSandboxID,
		ComponentID:      policyContext.ComponentID,
		ComponentBuildID: policyContext.ComponentBuildID,
		PolicyIDs:        policyIDs,
		InputCount:       len(policyInputs),
		OrgName:          policyContext.OrgName,
		AppName:          policyContext.AppName,
		InstallName:      policyContext.InstallName,
		ComponentName:    policyContext.ComponentName,
	}, nil
}

// policyContext is an alias for the shared PolicyEvaluationContext type.
type policyContext = policyhelpers.PolicyEvaluationContext

func (a *Activities) resolvePolicyContext(ctx context.Context, stepTargetID, stepTargetType string) (*policyContext, error) {
	switch app.WorkflowStepTargetType(stepTargetType) {
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		return a.resolveDeployPolicyContext(ctx, stepTargetID)
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		return a.resolveSandboxPolicyContext(ctx, stepTargetID)
	default:
		return nil, fmt.Errorf("unsupported step target type for policy checking: %s", stepTargetType)
	}
}

func (a *Activities) resolveDeployPolicyContext(ctx context.Context, deployID string) (*policyContext, error) {
	var deploy app.InstallDeploy
	res := a.db.WithContext(ctx).
		Preload("InstallComponent.Install.App.Org").
		Preload("InstallComponent.Component").
		First(&deploy, "id = ?", deployID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get deploy")
	}

	return &policyContext{
		AppConfigID:      deploy.InstallComponent.Install.AppConfigID,
		OrgID:            deploy.InstallComponent.Install.App.OrgID,
		AppID:            deploy.InstallComponent.Install.AppID,
		InstallID:        &deploy.InstallComponent.InstallID,
		ComponentID:      &deploy.InstallComponent.ComponentID,
		ComponentBuildID: &deploy.ComponentBuildID,
		ComponentType:    deploy.InstallComponent.Component.Type,
		ComponentName:    deploy.InstallComponent.Component.Name,
		IsSandbox:        false,
		OrgName:          deploy.InstallComponent.Install.App.Org.Name,
		AppName:          deploy.InstallComponent.Install.App.Name,
		InstallName:      deploy.InstallComponent.Install.Name,
	}, nil
}

func (a *Activities) resolveSandboxPolicyContext(ctx context.Context, sandboxRunID string) (*policyContext, error) {
	var sandboxRun app.InstallSandboxRun
	res := a.db.WithContext(ctx).
		Preload("Install.App.Org").
		First(&sandboxRun, "id = ?", sandboxRunID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get sandbox run")
	}

	return &policyContext{
		AppConfigID:      sandboxRun.Install.AppConfigID,
		OrgID:            sandboxRun.Install.App.OrgID,
		AppID:            sandboxRun.Install.AppID,
		InstallID:        &sandboxRun.InstallID,
		InstallSandboxID: sandboxRun.InstallSandboxID,
		ComponentType:    "",
		ComponentName:    "",
		IsSandbox:        true,
		OrgName:          sandboxRun.Install.App.Org.Name,
		AppName:          sandboxRun.Install.App.Name,
		InstallName:      sandboxRun.Install.Name,
	}, nil
}

func (a *Activities) filterApplicablePolicies(
	policies []app.AppPolicyConfig,
	componentType app.ComponentType,
	componentName string,
	isSandbox bool,
) []app.AppPolicyConfig {
	var applicable []app.AppPolicyConfig

	for _, policy := range policies {
		if policy.Engine != config.AppPolicyEngineOPA {
			continue
		}

		if isSandbox {
			if a.appliesForSandbox(policy) {
				applicable = append(applicable, policy)
			}
		} else {
			if a.appliesForComponent(policy, componentType, componentName) {
				applicable = append(applicable, policy)
			}
		}
	}

	return applicable
}

func (a *Activities) appliesForSandbox(policy app.AppPolicyConfig) bool {
	return policy.Type == config.AppPolicyTypeSandbox
}

func (a *Activities) appliesForComponent(
	policy app.AppPolicyConfig,
	componentType app.ComponentType,
	componentName string,
) bool {
	return policy.Type != config.AppPolicyTypeSandbox &&
		policyTypeMatchesComponentType(policy.Type, componentTypeToPolicyType(componentType)) &&
		len(policy.Components) > 0 &&
		(slices.Contains(policy.Components, componentName) || slices.Contains(policy.Components, "*"))
}

func policyTypeMatchesComponentType(policyType config.AppPolicyType, expectedType config.AppPolicyType) bool {
	return policyType == expectedType
}

func componentTypeToPolicyType(ct app.ComponentType) config.AppPolicyType {
	switch ct {
	case app.ComponentTypeTerraformModule:
		return config.AppPolicyTypeTerraformModule
	case app.ComponentTypeHelmChart:
		return config.AppPolicyTypeHelmChart
	case app.ComponentTypeKubernetesManifest:
		return config.AppPolicyTypeKubernetesManifest
	case app.ComponentTypeDockerBuild:
		return config.AppPolicyTypeDockerBuild
	case app.ComponentTypeExternalImage:
		return config.AppPolicyTypeContainerImage
	default:
		return ""
	}
}

func (a *Activities) preparePolicyInputs(planContentsJSON []byte, pctx *policyContext) ([][]byte, []string, error) {
	switch pctx.ComponentType {
	case app.ComponentTypeTerraformModule:
		return a.prepareTerraformPolicyInputs(planContentsJSON, pctx)
	case app.ComponentTypeHelmChart:
		return a.prepareHelmPolicyInputs(planContentsJSON)
	case app.ComponentTypeKubernetesManifest:
		return a.prepareKubernetesManifestPolicyInputs(planContentsJSON)
	default:
		return nil, nil, fmt.Errorf("unsupported component type for policy input preparation: %s", pctx.ComponentType)
	}
}

func (a *Activities) prepareTerraformPolicyInputs(planContentsJSON []byte, pctx *policyContext) ([][]byte, []string, error) {
	tfPlan, err := plan.ParseTerraformPlan(planContentsJSON)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse terraform plan")
	}

	planJSON, err := json.Marshal(tfPlan)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to marshal terraform plan for policy input")
	}

	identity := a.buildTerraformInputIdentity(pctx)
	return [][]byte{planJSON}, []string{identity}, nil
}

func (a *Activities) buildTerraformInputIdentity(pctx *policyContext) string {
	var parts []string

	if pctx.ComponentID != nil {
		componentPart := fmt.Sprintf("component:%s", *pctx.ComponentID)
		if pctx.ComponentName != "" {
			componentPart = fmt.Sprintf("component:%s (%s)", *pctx.ComponentID, pctx.ComponentName)
		}
		parts = append(parts, componentPart)
	}

	if pctx.ComponentBuildID != nil {
		parts = append(parts, fmt.Sprintf("build:%s", *pctx.ComponentBuildID))
	}

	if len(parts) == 0 {
		return "terraform-plan"
	}

	return fmt.Sprintf("%s", parts[0])
}

func (a *Activities) prepareKubernetesManifestPolicyInputs(planContentsJSON []byte) ([][]byte, []string, error) {
	var planContents plan.KubernetesManifestPlanContents
	if err := json.Unmarshal(planContentsJSON, &planContents); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal kubernetes manifest plan contents")
	}

	return a.yamlToAdmissionReviewInputs(planContents.DryRunOutput, planContentsJSON)
}

func (a *Activities) prepareHelmPolicyInputs(planContentsJSON []byte) ([][]byte, []string, error) {
	var planContents struct {
		TemplateOutput string `json:"template_output"`
	}
	if err := json.Unmarshal(planContentsJSON, &planContents); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal helm plan contents")
	}

	return a.yamlToAdmissionReviewInputs(planContents.TemplateOutput, planContentsJSON)
}

func (a *Activities) yamlToAdmissionReviewInputs(yamlManifests string, fallback []byte) ([][]byte, []string, error) {
	admissionReviews, err := plan.ParseMultiDocYAMLToAdmissionReviews(yamlManifests)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse manifests to admission reviews")
	}

	if len(admissionReviews) == 0 {
		return [][]byte{fallback}, []string{"unknown"}, nil
	}

	inputs := make([][]byte, len(admissionReviews))
	identities := make([]string, len(admissionReviews))
	for i, review := range admissionReviews {
		inputJSON, err := json.Marshal(review)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to marshal admission review %d", i)
		}
		inputs[i] = inputJSON
		identities[i] = extractKubernetesInputIdentity(review)
	}

	return inputs, identities, nil
}

func extractKubernetesInputIdentity(review plan.AdmissionReviewInput) string {
	kind := review.Review.Kind.Kind
	if kind == "" {
		kind = "Unknown"
	}

	namespace := ""
	name := ""

	if metadata, ok := review.Review.Object["metadata"].(map[string]interface{}); ok {
		if ns, ok := metadata["namespace"].(string); ok {
			namespace = ns
		}
		if n, ok := metadata["name"].(string); ok {
			name = n
		}
	}

	if namespace == "" {
		namespace = "default"
	}

	if name == "" {
		return fmt.Sprintf("%s/%s", kind, namespace)
	}

	return fmt.Sprintf("%s/%s/%s", kind, namespace, name)
}

func (a *Activities) buildPolicyEvaluationItems(
	policies []app.AppPolicyConfig,
	inputs [][]byte,
	identities []string,
) []PolicyToEvaluate {
	result := make([]PolicyToEvaluate, 0, len(policies)*len(inputs))

	for _, policy := range policies {
		for idx, input := range inputs {
			identity := ""
			if idx < len(identities) {
				identity = identities[idx]
			}
			result = append(result, PolicyToEvaluate{
				PolicyID:      policy.ID,
				PolicyName:    policy.Name,
				Contents:      policy.Contents,
				InputJSON:     input,
				InputIndex:    idx,
				InputIdentity: identity,
			})
		}
	}

	return result
}
