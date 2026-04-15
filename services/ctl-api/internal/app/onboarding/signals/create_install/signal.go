package createinstall

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "onboarding-create-install"

type Signal struct {
	signal.Hooks
	OnboardingID string                                      `json:"onboarding_id" validate:"required"`
	InstallName  string                                      `json:"install_name" validate:"required"`
	AWSAccount   *activities.CreateOnboardingInstallAWS      `json:"aws_account,omitempty"`
	AzureAccount *activities.CreateOnboardingInstallAzure    `json:"azure_account,omitempty"`
	Inputs       map[string]*string                          `json:"inputs,omitempty"`
	Metadata     *activities.CreateOnboardingInstallMetadata `json:"metadata,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	_, err := activities.AwaitGetOnboardingByOnboardingID(ctx, s.OnboardingID)
	if err != nil {
		return errors.Wrap(err, "onboarding not found")
	}

	return nil
}
