package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	kube "helm.sh/helm/v4/pkg/kube"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/diff"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) upgrade_diff(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config) (string, *[]diff.ResourceDiff, error) {
	l.Info("fetching previous release")
	prevRel, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if prevRel == nil {
		l.Warn("unable to fetch previous release, so assuming it failed and was not installed", zap.Error(err))
		l.Info("attempting install instead of upgrade")
		return h.installDiff(ctx, l, actionCfg, kubeCfg)
	}

	l.Info("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get chart")
	}

	l.Info("found default chart values", zap.Any("values", chart.Values))
	l.Info("loading provided values")
	values, err := helm.ChartValues(h.state.plan.HelmDeployPlan.ValuesFiles, h.state.plan.HelmDeployPlan.Values)
	if err != nil {
		return "", nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Info("rendered values", zap.Any("values", values))

	client := action.NewUpgrade(actionCfg)
	client.DryRun = true
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

	l.Info("calculating helm diff")
	rel, err := client.RunWithContext(ctx, prevRel.Name, chart, values)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}
	l.Info("parsing previous and current release manifests")
	diffH, diffReport, err := h.getDiff(l, kubeCfg, prevRel, rel, h.state.plan.HelmDeployPlan.Namespace)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}

	h.state.outputs = map[string]interface{}{"diff": diffH}

	contentDiff := make([]diff.ResourceDiff, 0)
	for _, re := range diffReport.Entries {
		d := diff.ResourceDiff{
			Version: "2",
		}

		namespace, name, kind, apiPath := diff.ParseRawResourceName(re.Key)
		d.Name = name
		d.Namespace = namespace
		d.Kind = kind
		d.ApiPath = apiPath

		for _, diffItem := range re.Diffs {
			d.Entries = append(d.Entries, diff.DiffEntry{
				Payload: diffItem.Payload,
				Type:    diff.DiffEntryType(diffItem.Delta),
			})
		}

		contentDiff = append(contentDiff, d)
	}

	return string(diffH), &contentDiff, nil
}

func (h *handler) installDiff(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config) (string, *[]diff.ResourceDiff, error) {
	l.Info("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get chart")
	}

	l.Info("found default chart values", zap.Any("values", chart.Values))

	l.Info("loading provided values")
	values, err := helm.ChartValues(h.state.plan.HelmDeployPlan.ValuesFiles, h.state.plan.HelmDeployPlan.Values)
	if err != nil {
		return "", nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Info("rendered values", zap.Any("values", values))

	client := action.NewInstall(actionCfg)
	client.ClientOnly = false
	client.DryRun = true
	client.DisableHooks = false

	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.Devel = true
	client.DependencyUpdate = true
	client.Timeout = h.state.timeout
	client.Namespace = h.state.plan.HelmDeployPlan.Namespace
	client.ReleaseName = h.state.plan.HelmDeployPlan.Name
	client.GenerateName = false
	client.NameTemplate = ""
	client.OutputDir = ""
	client.SkipCRDs = false
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Replace = false
	client.Description = ""
	client.CreateNamespace = h.state.plan.HelmDeployPlan.CreateNamespace

	l.Info("calculating helm diff", zap.String("operation", "diff"), zap.String("exec", "install"))
	rel, err := client.RunWithContext(ctx, chart, values)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}
	diffH, diffReport, err := h.getDiff(l, kubeCfg, nil, rel, h.state.plan.HelmDeployPlan.Namespace)

	h.state.outputs = map[string]interface{}{"diff": diffH}

	contentDiff := make([]diff.ResourceDiff, 0)
	for _, re := range diffReport.Entries {
		d := diff.ResourceDiff{
			Version: "2",
		}

		namespace, name, kind, apiPath := diff.ParseRawResourceName(re.Key)
		d.Name = name
		d.Namespace = namespace
		d.Kind = kind
		d.ApiPath = apiPath

		for _, diffItem := range re.Diffs {
			d.Entries = append(d.Entries, diff.DiffEntry{
				Payload: diffItem.Payload,
				Type:    diff.DiffEntryType(diffItem.Delta),
			})
		}

		contentDiff = append(contentDiff, d)
	}

	return string(diffH), &contentDiff, nil
}

func (h *handler) uninstallDiff(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config, prevRel *release.Release) (string, *[]diff.ResourceDiff, error) {
	// not functional atm (panics)
	l.Info("loading chart options")
	chart, err := helm.GetChartByPath(h.state.chartPath)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get chart")
	}

	l.Info("found default chart values", zap.Any("values", chart.Values))

	l.Info("loading provided values")
	values, err := helm.ChartValues(h.state.plan.HelmDeployPlan.ValuesFiles, h.state.plan.HelmDeployPlan.Values)
	if err != nil {
		return "", nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Info("rendered values", zap.Any("values", values))

	client := action.NewUninstall(actionCfg)
	client.DryRun = true
	client.DisableHooks = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.Timeout = h.state.timeout

	l.Info("calculating helm diff", zap.String("operation", "diff"), zap.String("exec", "uninstall"))
	resp, err := client.Run(prevRel.Name)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}
	l.Info(resp.Info)

	diffH, diffReport, err := h.getDiff(l, kubeCfg, prevRel, nil, h.state.plan.HelmDeployPlan.Namespace)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}

	h.state.outputs = map[string]interface{}{"diff": diffH}

	contentDiff := make([]diff.ResourceDiff, 0)
	for _, re := range diffReport.Entries {
		d := diff.ResourceDiff{
			Version: "2",
		}

		namespace, name, kind, apiPath := diff.ParseRawResourceName(re.Key)
		d.Name = name
		d.Namespace = namespace
		d.Kind = kind
		d.ApiPath = apiPath

		for _, diffItem := range re.Diffs {
			d.Entries = append(d.Entries, diff.DiffEntry{
				Payload: diffItem.Payload,
				Type:    diff.DiffEntryType(diffItem.Delta),
			})
		}

		contentDiff = append(contentDiff, d)
	}

	return string(diffH), &contentDiff, nil
}
