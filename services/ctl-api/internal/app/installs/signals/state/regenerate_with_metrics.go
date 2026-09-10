package state

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

type RegenerateWithMetricsRequest struct {
	InstallID       string
	Targets         []pkgstate.PartialTarget
	AllTargets      bool
	ForceAll        bool
	TriggeredByID   string
	TriggeredByType string
	MetricsWriter   metrics.Writer
}

// RegenerateWithMetrics runs Regenerate and emits the nuon.state.regenerate.*
// series. Callers that bypass the state-partial-generate signal must use this
// rather than Regenerate directly, or those metrics silently stop being emitted
// for their path.
func RegenerateWithMetrics(ctx workflow.Context, req RegenerateWithMetricsRequest) error {
	targets := req.Targets
	if req.AllTargets {
		targets = pkgstate.AllPartialTargets()
	}

	start := time.Now()
	resp, err := Regenerate(ctx, &pkgstate.ExecuteRegenerationRequest{
		InstallID:       req.InstallID,
		Targets:         targets,
		TriggeredByID:   req.TriggeredByID,
		TriggeredByType: req.TriggeredByType,
		MetricsWriter:   req.MetricsWriter,
	})
	if err != nil {
		return err
	}
	runtime := time.Since(start)

	if req.MetricsWriter == nil {
		return nil
	}

	appID, appName, appConfigID := "", "", ""
	if resp != nil {
		appID = resp.AppID
		appName = resp.AppName
		appConfigID = resp.AppConfigID
	}
	baseTags := metrics.ToTags(map[string]string{
		"install_id":        req.InstallID,
		"triggered_by_type": req.TriggeredByType,
		"app_id":            appID,
		"app_name":          appName,
		"app_config_id":     appConfigID,
	})

	req.MetricsWriter.Timing("nuon.state.regenerate.duration", runtime, baseTags)
	if req.AllTargets {
		req.MetricsWriter.Timing("nuon.state.regenerate.full.duration", runtime, baseTags)
		req.MetricsWriter.Count("nuon.state.regenerate.full.count", 1, baseTags)
	} else {
		req.MetricsWriter.Timing("nuon.state.regenerate.partial.duration", runtime, baseTags)
		req.MetricsWriter.Count("nuon.state.regenerate.partial.count", 1, baseTags)
	}

	if resp == nil {
		return nil
	}
	for _, partial := range resp.UpdatedPartials {
		partialTags := metrics.ToTags(map[string]string{
			"install_id":        req.InstallID,
			"partial":           string(partial),
			"triggered_by_type": req.TriggeredByType,
			"app_id":            appID,
			"app_name":          appName,
			"app_config_id":     appConfigID,
		})
		req.MetricsWriter.Count("nuon.state.regenerate.partial.updated", 1, partialTags)
	}

	return nil
}
