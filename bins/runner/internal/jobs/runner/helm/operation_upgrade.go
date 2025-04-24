package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/powertoolsdev/mono/pkg/helm"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func (h *handler) upgrade(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*release.Release, error) {
	l.Info("fetching previous release")
	prevRel, err := helm.GetRelease(actionCfg, h.state.cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("unable to get previous helm release: %w", err)
	}

	l.Info("loading chart", zap.String("repo", h.state.cfg.Repository))
	c, err := helm.GetChartByPath(h.state.cfg.Repository)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart: %w", err)
	}

	l.Info("found default chart values", zap.Any("values", c.Values))

	l.Info("loading provided values")
	vals := make([]plantypes.HelmValue, 0)
	for _, val := range h.state.cfg.HelmSet {
		vals = append(vals, plantypes.HelmValue{
			Name:  val.Name,
			Value: val.Value,
		})
	}
	values, err := helm.ChartValues(h.state.cfg.Values, vals)

	// We have a previous release, upgrade.
	client := action.NewUpgrade(actionCfg)
	client.DryRun = false
	client.DisableHooks = false
	client.Wait = true
	client.WaitForJobs = false
	client.Devel = h.state.cfg.Devel
	client.DependencyUpdate = true
	client.Timeout = h.state.timeout
	client.Namespace = h.state.cfg.Namespace
	client.Atomic = false
	client.SkipCRDs = h.state.cfg.SkipCRDs
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Description = ""
	client.ResetValues = false
	client.ReuseValues = false
	client.Recreate = false
	client.MaxHistory = 0
	client.CleanupOnFail = false
	client.Force = false

	l.Info("upgrading helm release")
	rel, err := client.RunWithContext(ctx, prevRel.Name, c, values)
	if err != nil {
		return nil, fmt.Errorf("unable to upgrade helm release: %w", err)
	}

	return rel, nil
}
