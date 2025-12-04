package activities

import (
	"context"
	"fmt"
	"slices"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type PrepPolicyEvaluationRequest struct {
	StepTargetID   string `validate:"required"`
	StepTargetType string `validate:"required"`
}

type PolicyViolation struct {
	PolicyID string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	Message  string `json:"message" temporaljson:"message,omitempty"`
}

type PolicyToEvaluate struct {
	PolicyID string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	Contents string `json:"contents" temporaljson:"contents,omitempty"`
}

type PrepPolicyEvaluationResult struct {
	Policies    []PolicyToEvaluate `json:"policies" temporaljson:"policies,omitempty"`
	InputJSON   []byte             `json:"input_json" temporaljson:"input_json,omitempty"`
	HasPolicies bool               `json:"has_policies" temporaljson:"has_policies,omitempty"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) PrepPolicyEvaluation(ctx context.Context, req *PrepPolicyEvaluationRequest) (*PrepPolicyEvaluationResult, error) {
	policyContext, err := a.resolvePolicyContext(ctx, req.StepTargetID, req.StepTargetType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve policy context")
	}

	policiesConfig, err := a.getPoliciesConfigByAppConfigID(ctx, policyContext.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get policies config")
	}

	plan, err := a.getApprovalPlan(ctx, req.StepTargetID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get plan contents")
	}

	applicablePolicies := a.filterApplicablePolicies(
		policiesConfig.Policies,
		policyContext.ComponentType,
		policyContext.ComponentName,
		policyContext.IsSandbox,
	)

	if len(applicablePolicies) == 0 {
		return &PrepPolicyEvaluationResult{
			Policies:    []PolicyToEvaluate{},
			HasPolicies: false,
		}, nil
	}

	policyInput, err := a.preparePolicyInput(plan.PlanContents, policyContext.ComponentType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare policy input")
	}

	policies := make([]PolicyToEvaluate, len(applicablePolicies))
	for i, p := range applicablePolicies {
		policies[i] = PolicyToEvaluate{
			PolicyID: p.ID,
			Contents: p.Contents,
		}
	}

	return &PrepPolicyEvaluationResult{
		Policies:    policies,
		InputJSON:   policyInput,
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
	switch stepTargetType {
	case app.WorkflowStepTargetTypeInstallDeploy:
		return a.resolveDeployPolicyContext(ctx, stepTargetID)
	case app.WorkflowStepTargetTypeInstallSandboxRun:
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
	return policy.Sandbox &&
		policyTypeMatchesComponentType(policy.Type, config.AppPolicyTypeTerraformModule)
}

func (a *Activities) appliesForComponent(
	policy app.AppPolicyConfig,
	componentType app.ComponentType,
	componentName string,
) bool {
	return !policy.Sandbox &&
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

func (a *Activities) preparePolicyInput(inputJSON []byte, componentType app.ComponentType) ([]byte, error) {
	switch componentType {
	case app.ComponentTypeTerraformModule:
		return inputJSON, nil
	case app.ComponentTypeHelmChart:
		return inputJSON, nil
	case app.ComponentTypeKubernetesManifest:
		return inputJSON, nil
	default:
		return nil, fmt.Errorf("unsupported component type for policy input preparation: %s", componentType)
	}
}
