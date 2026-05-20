package slackautolink

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	generalactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "general-slack-auto-link"

var _ signal.Signal = (*Signal)(nil)

// Signal has no per-invocation params — the activity reads cfg.SlackAutoLink*.
type Signal struct{}

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error { return nil }

func (s *Signal) Execute(ctx workflow.Context) error {
	if _, err := generalactivities.AwaitEnsureSlackAutoLinks(ctx, generalactivities.EnsureSlackAutoLinksRequest{}); err != nil {
		return fmt.Errorf("ensure slack auto links: %w", err)
	}
	return nil
}
