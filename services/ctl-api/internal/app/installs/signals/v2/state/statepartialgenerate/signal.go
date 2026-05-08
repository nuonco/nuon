package statepartialgenerate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/metrics"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
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
	start := time.Now()
	if s.AllTargets {
		s.Targets = pkgstate.AllPartialTargets()
	}
	resp, err := statesignals.Regenerate(ctx, &state.ExecuteRegenerationRequest{
		InstallID:       s.InstallID,
		Targets:         s.Targets,
		TriggeredByID:   s.TriggeredByID,
		TriggeredByType: s.TriggeredByType,
		MetricsWriter:   s.metrics,
	})
	if err != nil {
		return errors.Wrap(err, "unable to regenerate state")
	}
	runtime := time.Since(start)

	// emit metrics
	if s.metrics == nil {
		return nil
	}
	appID, appName, appConfigID := "", "", ""
	if resp != nil {
		appID = resp.AppID
		appName = resp.AppName
		appConfigID = resp.AppConfigID
	}
	baseTags := metrics.ToTags(map[string]string{
		"install_id":        s.InstallID,
		"triggered_by_type": s.TriggeredByType,
		"app_id":            appID,
		"app_name":          appName,
		"app_config_id":     appConfigID,
	})
	s.metrics.Timing("nuon.state.regenerate.duration", runtime, baseTags)
	if s.AllTargets {
		s.metrics.Timing("nuon.state.regenerate.full.duration", runtime, baseTags)
		s.metrics.Count("nuon.state.regenerate.full.count", 1, baseTags)
	} else {
		s.metrics.Timing("nuon.state.regenerate.partial.duration", runtime, baseTags)
		s.metrics.Count("nuon.state.regenerate.partial.count", 1, baseTags)
	}

	if resp != nil {
		for _, partial := range resp.UpdatedPartials {
			partialTags := metrics.ToTags(map[string]string{
				"install_id":        s.InstallID,
				"partial":           string(partial),
				"triggered_by_type": s.TriggeredByType,
				"app_id":            appID,
				"app_name":          appName,
				"app_config_id":     appConfigID,
			})
			s.metrics.Count("nuon.state.regenerate.partial.updated", 1, partialTags)
		}
	}

	return err
}
