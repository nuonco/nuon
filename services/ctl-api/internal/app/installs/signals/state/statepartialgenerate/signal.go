package statepartialgenerate

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/metrics"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

const SignalType signal.SignalType = "state-partial-generate"

// Signal generates state based on signal input
type Signal struct {
	InstallID       string
	Targets         []state.PartialTarget
	AllTargets      bool
	ForceAll        bool
	TriggeredByID   string
	TriggeredByType string

	v       *validator.Validate
	metrics metrics.Writer
}

var (
	_ signal.Signal              = &Signal{}
	_ signal.SignalWithAutoRetry = (*Signal)(nil)
	_ signal.SignalWithParams    = (*Signal)(nil)
)

func (s *Signal) WithParams(params *signal.Params) {
	s.v = params.V
	s.metrics = params.MW
}

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if len(s.Targets) == 0 && !s.AllTargets {
		return fmt.Errorf("targets is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return statesignals.RegenerateWithMetrics(ctx, statesignals.RegenerateWithMetricsRequest{
		InstallID:       s.InstallID,
		Targets:         s.Targets,
		AllTargets:      s.AllTargets,
		ForceAll:        s.ForceAll,
		TriggeredByID:   s.TriggeredByID,
		TriggeredByType: s.TriggeredByType,
		MetricsWriter:   s.metrics,
	})
}
