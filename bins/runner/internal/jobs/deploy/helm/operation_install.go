package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/outputs"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) install(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) (*release.Release, error) {
	l.Info("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get chart")
	}

	l.Info("loading chart values")
	values, err := helm.ChartValues(h.state.cfg.Values, h.state.cfg.HelmSet)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}

	client := action.NewInstall(actionCfg)
	client.ClientOnly = false
	client.DryRun = true
	client.DisableHooks = false
	client.Wait = true
	client.WaitForJobs = false
	client.Devel = h.state.cfg.Devel
	client.DependencyUpdate = true
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

	l.Info("calculating helm diff")
	rel, err := client.RunWithContext(ctx, chart, values)
	if err != nil {
		return nil, errors.Wrap(err, "unable to execute with dry-run")
	}
	newMapping := manifest.Parse(rel.Manifest, rel.Namespace, true)
	if err := h.logDiff(l, map[string]*manifest.MappingResult{}, newMapping); err != nil {
		return nil, errors.Wrap(err, "unable to execute with dry-run")
	}

	l.Info("running helm install")
	client.DryRun = false
	rel, err = client.RunWithContext(ctx, chart, values)
	if err != nil {
		return nil, fmt.Errorf("unable to install chart: %w", err)
	}

	// NOTE(jm): we parse these here, so we have more context and the hanging action client, vs passing more stuff around.
	outs, err := outputs.HelmOutputs(rel.Manifest, rel.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse outputs")
	}
	h.state.outputs = outs

	return rel, nil
}
