package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateOnboardingInput struct {
	OnboardingID string  `json:"onboarding_id" validate:"required"`
	OrgID        *string `json:"org_id,omitempty"`
	AppID        *string `json:"app_id,omitempty"`
	AppBranchID  *string `json:"app_branch_id,omitempty"`
	InstallID    *string `json:"install_id,omitempty"`
	WorkflowID   *string `json:"workflow_id,omitempty"`
	InstallMode  *string `json:"install_mode,omitempty"`
	CurrentStep  *string `json:"current_step,omitempty"`
	StepStatus   *string `json:"step_status,omitempty"`
	StepError    *string `json:"step_error,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) updateOnboarding(ctx context.Context, req *UpdateOnboardingInput) (*app.Onboarding, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var onboarding app.Onboarding
	if err := a.db.WithContext(ctx).First(&onboarding, "id = ?", req.OnboardingID).Error; err != nil {
		return nil, fmt.Errorf("onboarding not found: %w", err)
	}

	if req.OrgID != nil {
		onboarding.OrgID = req.OrgID
	}
	if req.AppID != nil {
		onboarding.AppID = req.AppID
	}
	if req.AppBranchID != nil {
		onboarding.AppBranchID = req.AppBranchID
	}
	if req.InstallID != nil {
		onboarding.InstallID = req.InstallID
	}
	if req.WorkflowID != nil {
		onboarding.WorkflowID = req.WorkflowID
	}
	if req.InstallMode != nil {
		onboarding.InstallMode = app.OnboardingInstallMode(*req.InstallMode)
	}
	if req.CurrentStep != nil {
		onboarding.CurrentStep = app.OnboardingStep(*req.CurrentStep)
	}
	if req.StepStatus != nil {
		onboarding.StepStatus = app.OnboardingStepStatus(*req.StepStatus)
	}
	if req.StepError != nil {
		onboarding.StepError = req.StepError
	}
	// Clear step error when status is not error
	if req.StepStatus != nil && app.OnboardingStepStatus(*req.StepStatus) != app.OnboardingStepStatusError {
		onboarding.StepError = nil
	}

	if err := a.db.WithContext(ctx).Save(&onboarding).Error; err != nil {
		return nil, fmt.Errorf("unable to update onboarding: %w", err)
	}

	return &onboarding, nil
}
