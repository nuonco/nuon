package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/rest"

	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/outputs"
	"github.com/nuonco/nuon/pkg/helm"
)

func (h *handler) install(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config) (*release.Release, error) {
	l.Debug("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get chart")
	}

	l = l.With(zap.String("helm.chart_name", chart.Name()))

	l.Debug("found default chart values", zap.Any("values", chart.Values))

	l.Debug("loading provided values")
	values, err := helm.ChartValues(h.state.plan.HelmDeployPlan.ValuesFiles, h.state.plan.HelmDeployPlan.Values, h.state.plan.HelmDeployPlan.ValuesOverride)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Debug("rendered values", zap.Any("values", values))

	// get the default client
	client := helm.DefaultInstall(actionCfg)
	// configure the client with additional, chart and context -specific values
	client.CreateNamespace = h.state.plan.HelmDeployPlan.CreateNamespace
	client.Namespace = h.state.plan.HelmDeployPlan.Namespace
	client.ReleaseName = h.state.plan.HelmDeployPlan.Name
	client.TakeOwnership = h.state.plan.HelmDeployPlan.TakeOwnership
	client.SkipCRDs = h.state.plan.HelmDeployPlan.SkipCRDs
	client.Timeout = h.state.timeout
	client.DryRun = true

	// determine if we're going to calculate the diff
	var rendered string
	crds := chart.CRDObjects()
	if len(crds) > 0 && !client.SkipCRDs {
		// skip dry run
		crdZapFieldList := []zap.Field{}
		for i, crd := range crds {
			field := zap.String(fmt.Sprintf("crd.%d", i), crd.Name)
			crdZapFieldList = append(crdZapFieldList, field)
		}
		l.Info(
			"chart contains CRDs - skipping dry-run",
			crdZapFieldList...,
		)
		rendered, err = helm.RenderManifest(chart, values, client.ReleaseName, client.Namespace)
		if err != nil {
			return nil, fmt.Errorf("unable to render helm chart: %w", err)
		}
	} else {
		l.Info("calculating helm diff", zap.String("operation", "install"), zap.String("exec", "install"))
		rel, err := client.RunWithContext(ctx, chart, values)
		if err != nil {
			return nil, errors.Wrap(err, "unable to execute with dry-run")
		}
		rendered = rel.Manifest
		newMapping := manifest.Parse(rel.Manifest, rel.Namespace, true)
		if err := h.logDiff(l, map[string]*manifest.MappingResult{}, newMapping); err != nil {
			return nil, errors.Wrap(err, "unable to execute with dry-run")
		}

	}

	resources, err := helm.ResourcesFromManifest(rendered)
	if err != nil {
		return nil, fmt.Errorf("unable to read rendered chart: %w", err)
	}
	if err := helm.CheckResourceAccess(ctx, kubeCfg, resources, client.Namespace, l); err != nil {
		return nil, fmt.Errorf("unable to access chart resources: %w", err)
	}

	l.Info("running helm install")
	client.DryRun = false
	rel, err := helm.HelmInstallWithLogStreaming(ctx, client, chart, values, kubeCfg, l)
	if err != nil {
		return nil, fmt.Errorf("unable to upgrade helm release: %w", err)
	}

	// NOTE(jm): we parse these here, so we have more context and the hanging action client, vs passing more stuff around.
	outs, err := outputs.HelmOutputs(rel.Manifest, rel.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse outputs")
	}

	live, err := helm.CollectLiveOutputs(ctx, kubeCfg, resources, rel.Namespace, l)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve chart resources from kubernetes: %w", err)
	}
	outs["ingresses"] = live.Ingresses
	outs["services"] = live.Services
	outs["deployments"] = live.Deployments
	h.state.outputs = outs

	return rel, nil
}
