package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) upgrade(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*release.Release, error) {
	l.Info("fetching previous release")
	prevRel, err := helm.GetRelease(actionCfg, h.state.cfg.Name)
	if prevRel == nil {
		l.Warn("unable to fetch previous release, so assuming it failed and was not installed", zap.Error(err))
		l.Info("attempting install instead of upgrade")
		return h.install(ctx, l, actionCfg)
	}

	l.Info("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get chart")
	}

	values, err := helm.ChartValues(h.state.cfg.Values, h.state.cfg.HelmSet)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}

	// We have a previous release, upgrade.
	client := action.NewUpgrade(actionCfg)
	client.DryRun = false
	client.DisableHooks = false
	client.Wait = true
	client.WaitForJobs = false
	client.Devel = h.state.cfg.Devel
	client.DependencyUpdate = false
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
	rel, err := client.RunWithContext(ctx, prevRel.Name, chart, values)
	if err != nil {
		return nil, fmt.Errorf("unable to upgrade helm release: %w", err)
	}

	return rel, nil
}
