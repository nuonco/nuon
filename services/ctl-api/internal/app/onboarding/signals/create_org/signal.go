package createorg

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "onboarding-create-org"

type Signal struct {
	OnboardingID string `json:"onboarding_id" validate:"required"`
	AccountID    string `json:"account_id" validate:"required"`
	OrgName      string `json:"org_name" validate:"required"`
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
