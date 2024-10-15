package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) install(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*release.Release, error) {
	l.Info("loading helm env settings")
	settings, err := helm.LoadEnvSettings()
	if err != nil {
		return nil, fmt.Errorf("unable to load env settings: %w", err)
	}

	l.Info("loading chart options")
	cpo, chartName, err := helm.ChartPathOptions(
		h.state.cfg.Repository,
		h.state.cfg.Chart,
		h.state.cfg.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart options: %w", err)
	}

	l.Info("loading chart")
	c, _, err := helm.GetChart(chartName, cpo, settings)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart: %w", err)
	}

	l.Info("loading chart values")
	values, err := helm.ChartValues(h.state.cfg.Values, h.state.cfg.HelmSet)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}

	client := action.NewInstall(actionCfg)
	client.ChartPathOptions = *cpo
	client.ClientOnly = false
	client.DryRun = false
	client.DisableHooks = false
	client.Wait = true
	client.WaitForJobs = false
	client.Devel = h.state.cfg.Devel
	client.DependencyUpdate = false
	client.Timeout = h.state.timeout
	client.Namespace = h.state.cfg.Namespace
	client.ReleaseName = h.state.cfg.Name
	client.GenerateName = false
	client.NameTemplate = ""
	client.OutputDir = ""
	client.Atomic = false
	client.SkipCRDs = h.state.cfg.SkipCRDs
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Replace = false
	client.Description = ""
	client.CreateNamespace = h.state.cfg.CreateNamespace

	l.Info("running install")
	rel, err := client.Run(c, values)
	if err != nil {
		return nil, fmt.Errorf("unable to install chart: %w", err)
	}

	return rel, nil
}
