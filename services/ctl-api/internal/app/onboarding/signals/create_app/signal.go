package createapp

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "onboarding-create-app"

type Signal struct {
	signal.Hooks
	OnboardingID string `json:"onboarding_id" validate:"required"`

	// Example app fields (populated when AppType=example)
	ExampleRepo      string `json:"example_repo,omitempty"`
	ExampleDirectory string `json:"example_directory,omitempty"`
	ExampleBranch    string `json:"example_branch,omitempty"`
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
