package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/helm"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func (h *handler) install(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*release.Release, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Info("parsed values", zap.Any("values", values))

	client := action.NewInstall(actionCfg)
	client.ClientOnly = false
	client.DryRun = false
	client.DisableHooks = false
	client.Wait = true
	client.WaitForJobs = false
	client.Devel = h.state.cfg.Devel
	client.DependencyUpdate = true
	client.Timeout = h.state.timeout
	client.Namespace = h.state.cfg.Namespace
	client.ReleaseName = h.state.cfg.Name
	client.GenerateName = false
	client.ReleaseName = h.state.cfg.Name
	client.OutputDir = ""
	client.Atomic = false
	client.SkipCRDs = h.state.cfg.SkipCRDs
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Replace = false
	client.Description = ""
	client.CreateNamespace = h.state.cfg.CreateNamespace

	l.Info("running install")
	rel, err := client.RunWithContext(ctx, c, values)
	if err != nil {
		return nil, fmt.Errorf("unable to install chart: %w", err)
	}

	return rel, nil
}
