package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	kube "helm.sh/helm/v4/pkg/kube"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/rest"

	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/outputs"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) upgrade(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config) (*release.Release, error) {
	l.Info("fetching previous release")
	prevRel, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if prevRel == nil {
		l.Warn("unable to fetch previous release, so assuming it failed and was not installed", zap.Error(err))
		l.Info("attempting install instead of upgrade")
		return h.install(ctx, l, actionCfg, kubeCfg)
	}
	l.Info(
		"found previous release",
		zap.String("name", prevRel.Name),
		zap.String("namespace", prevRel.Namespace),
		zap.Int("version", prevRel.Version),
		zap.Time("first_deployed", prevRel.Info.FirstDeployed.Time),
		zap.Time("last_deployed", prevRel.Info.LastDeployed.Time),
	)

	l.Info("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get chart")
	}

	l.Info("found default chart values", zap.Any("values", chart.Values))
	l.Info("loading provided values")
	values, err := helm.ChartValues(h.state.plan.HelmDeployPlan.ValuesFiles, h.state.plan.HelmDeployPlan.Values)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Info("rendered values", zap.Any("values", values))

	client := action.NewUpgrade(actionCfg)
	client.DryRun = true
	client.DisableHooks = false
	// wait logic
	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.Devel = true
	client.DependencyUpdate = true
	client.Timeout = h.state.timeout
	client.Namespace = h.state.plan.HelmDeployPlan.Namespace
	client.SkipCRDs = false
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Description = ""
	client.ResetValues = false
	client.ReuseValues = false
	client.MaxHistory = 0
	client.CleanupOnFail = false
	client.TakeOwnership = h.state.plan.HelmDeployPlan.TakeOwnership

	crds := chart.CRDObjects()
	if len(crds) > 0 {
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
	} else {

		l.Info("calculating helm diff")
		rel, err := client.RunWithContext(ctx, prevRel.Name, chart, values)
		if err != nil {
			return nil, errors.Wrap(err, "unable to execute with dry-run")
		}
		prevMapping := manifest.Parse(prevRel.Manifest, prevRel.Namespace, true)
		newMapping := manifest.Parse(rel.Manifest, rel.Namespace, true)
		if err := h.logDiff(l, prevMapping, newMapping); err != nil {
			return nil, errors.Wrap(err, "unable to execute with dry-run")
		}
	}

	l.Info("upgrading helm release")
	client = action.NewUpgrade(actionCfg)
	client.DryRun = false
	client.DisableHooks = false

	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy

	client.Devel = true
	client.DependencyUpdate = true
	client.Timeout = h.state.timeout
	client.Namespace = h.state.plan.HelmDeployPlan.Namespace
	client.SkipCRDs = false
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Description = ""
	client.ResetValues = false
	client.ReuseValues = false
	client.MaxHistory = 0
	client.CleanupOnFail = false
	client.TakeOwnership = h.state.plan.HelmDeployPlan.TakeOwnership

	rel, err := helm.HelmUpgradeWithLogStreaming(ctx, client, prevRel.Name, chart, values, kubeCfg, l)
	if err != nil {
		return nil, fmt.Errorf("unable to upgrade helm release: %w", err)
	}

	// NOTE(jm): we parse these here, so we have more context and the hanging action client, vs passing more stuff
	// around.
	outs, err := outputs.HelmOutputs(rel.Manifest, rel.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse outputs")
	}

	ingressOutputs, err := outputs.K8SGetHelmReleaseIngresses(ctx, rel.Name, kubeCfg, l)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve ingresses for this release from k8s")
	}
	serviceOutputs, err := outputs.K8SGetHelmReleaseServices(ctx, rel.Name, kubeCfg, l)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve services for this release from k8s")
	}
	deploymentOutputs, err := outputs.K8SGetHelmReleaseDeployments(ctx, rel.Name, kubeCfg, l)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve deployments for this release from k8s")
	}

	outs["ingresses"] = ingressOutputs
	outs["services"] = serviceOutputs
	outs["deployments"] = deploymentOutputs
	h.state.outputs = outs

	return rel, nil
}
