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
)

type PrepPolicyEvaluationRequest struct {
	StepTargetID   string `validate:"required"`
	StepTargetType string `validate:"required"`
}

type PolicyViolation struct {
	PolicyID string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	Message  string `json:"message" temporaljson:"message,omitempty"`
	Severity string `json:"severity" temporaljson:"severity,omitempty"` // "deny" or "warn"
}

type PolicyToEvaluate struct {
	PolicyID  string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	Contents  string `json:"contents" temporaljson:"contents,omitempty"`
	InputJSON []byte `json:"input_json" temporaljson:"input_json,omitempty"`
}

type PrepPolicyEvaluationResult struct {
	Policies    []PolicyToEvaluate `json:"policies" temporaljson:"policies,omitempty"`
	HasPolicies bool               `json:"has_policies" temporaljson:"has_policies,omitempty"`
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

	policiesConfig, err := a.getPoliciesConfigByAppConfigID(ctx, policyContext.AppConfigID)
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
			Policies:    []PolicyToEvaluate{},
			HasPolicies: false,
		}, nil
	}

	policyInputs, err := a.preparePolicyInputs(approvalPlan.PlanContents, policyContext.ComponentType)
	if err != nil {
		l.Error("unable to prepare policy inputs", zap.Error(err))
		return nil, errors.Wrap(err, "unable to prepare policy inputs")
	}

	policies := a.buildPolicyEvaluationItems(applicablePolicies, policyInputs)

	l.Info("policy evaluation preparation complete",
		zap.Int("policies_count", len(applicablePolicies)),
		zap.Int("inputs_count", len(policyInputs)),
		zap.Int("total_evaluations", len(policies)),
	)

	return &PrepPolicyEvaluationResult{
		Policies:    policies,
		HasPolicies: true,
	}, nil
}

type policyContext struct {
	AppConfigID   string
	ComponentType app.ComponentType
	ComponentName string
	IsSandbox     bool
}

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
		Preload("InstallComponent.Install").
		Preload("InstallComponent.Component").
		First(&deploy, "id = ?", deployID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get deploy")
	}

	return &policyContext{
		AppConfigID:   deploy.InstallComponent.Install.AppConfigID,
		ComponentType: deploy.InstallComponent.Component.Type,
		ComponentName: deploy.InstallComponent.Component.Name,
		IsSandbox:     false,
	}, nil
}

func (a *Activities) resolveSandboxPolicyContext(ctx context.Context, sandboxRunID string) (*policyContext, error) {
	var sandboxRun app.InstallSandboxRun
	res := a.db.WithContext(ctx).
		Preload("Install").
		First(&sandboxRun, "id = ?", sandboxRunID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get sandbox run")
	}

	return &policyContext{
		AppConfigID:   sandboxRun.Install.AppConfigID,
		ComponentType: "",
		ComponentName: "",
		IsSandbox:     true,
	}, nil
}

func (a *Activities) getPoliciesConfigByAppConfigID(ctx context.Context, appConfigID string) (*app.AppPoliciesConfig, error) {
	var policiesConfig app.AppPoliciesConfig
	res := a.db.WithContext(ctx).
		Where("app_config_id = ?", appConfigID).
		Preload("Policies").
		Order("created_at DESC").
		First(&policiesConfig)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get policies config")
	}
	return &policiesConfig, nil
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

func (a *Activities) preparePolicyInputs(planContentsJSON []byte, componentType app.ComponentType) ([][]byte, error) {
	switch componentType {
	case app.ComponentTypeTerraformModule:
		return [][]byte{planContentsJSON}, nil
	case app.ComponentTypeHelmChart:
		return [][]byte{planContentsJSON}, nil
	case app.ComponentTypeKubernetesManifest:
		return a.prepareKubernetesManifestPolicyInputs(planContentsJSON)
	default:
		return nil, fmt.Errorf("unsupported component type for policy input preparation: %s", componentType)
	}
}

func (a *Activities) prepareKubernetesManifestPolicyInputs(planContentsJSON []byte) ([][]byte, error) {
	var planContents plan.KubernetesManifestPlanContents
	if err := json.Unmarshal(planContentsJSON, &planContents); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal kubernetes manifest plan contents")
	}

	admissionReviews, err := plan.ParseMultiDocYAMLToAdmissionReviews(planContents.DryRunOutput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dry run output to admission reviews")
	}

	if len(admissionReviews) == 0 {
		return [][]byte{planContentsJSON}, nil
	}

	inputs := make([][]byte, len(admissionReviews))
	for i, review := range admissionReviews {
		inputJSON, err := json.Marshal(review)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal admission review %d", i)
		}
		inputs[i] = inputJSON
	}

	return inputs, nil
}

func (a *Activities) buildPolicyEvaluationItems(
	policies []app.AppPolicyConfig,
	inputs [][]byte,
) []PolicyToEvaluate {
	result := make([]PolicyToEvaluate, 0, len(policies)*len(inputs))

	for _, policy := range policies {
		for _, input := range inputs {
			result = append(result, PolicyToEvaluate{
				PolicyID:  policy.ID,
				Contents:  policy.Contents,
				InputJSON: input,
			})
		}
	}

	return result
}
