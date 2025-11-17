package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	helmkube "helm.sh/helm/v4/pkg/kube"
	release "helm.sh/helm/v4/pkg/release/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
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
	client.WaitStrategy = helmkube.StatusWatcherStrategy
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

func applyCRDs(ctx context.Context, l *zap.Logger, kubeCfg *rest.Config, crdYAMLs [][]byte) error {
	dynClient, err := dynamic.NewForConfig(kubeCfg)
	if err != nil {
		return errors.Wrap(err, "unable to create dynamic client for CRDs")
	}

	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	for i, crdYAML := range crdYAMLs {
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal(crdYAML, &obj); err != nil {
			l.Warn("unable to parse CRD YAML", zap.Int("index", i), zap.Error(err))
			continue
		}

		l.Debug("Applying CRD", zap.Int("index", i), zap.String("name", obj.GetName()))

		// Apply to cluster
		applyOptions := metav1.ApplyOptions{
			FieldManager: "helm-crd-installer",
			Force:        true,
		}

		_, err := dynClient.Resource(crdGVR).Apply(ctx, obj.GetName(), &obj, applyOptions)
		if err != nil {
			return errors.Wrapf(err, "unable to apply CRD at index %d", i)
		}
	}

	return nil
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
	client.WaitStrategy = helmkube.StatusWatcherStrategy
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

	crds := chart.CRDObjects()
	if len(crds) > 0 {
		crdZapFieldList := []zap.Field{}
		crdYAMLs := make([][]byte, 0, len(crds))

		for i, crd := range crds {
			field := zap.String(fmt.Sprintf("crd.%d", i), crd.Name)
			crdZapFieldList = append(crdZapFieldList, field)
			if crd.File != nil {
				crdYAMLs = append(crdYAMLs, crd.File.Data)
			}
		}

		l.Info("chart contains CRDs - installing them for dry-run", crdZapFieldList...)

		// Apply CRDs to the cluster so dry-run can validate against the CRD types
		if err := applyCRDs(ctx, l, kubeCfg, crdYAMLs); err != nil {
			l.Warn("failed to apply CRDs for dry-run, continuing anyway", zap.Error(err))
		}

		client.SkipCRDs = true
	}

	l.Info("calculating helm diff", zap.String("operation", "diff"), zap.String("exec", "install"))
	rel, err := client.RunWithContext(ctx, chart, values)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}
	diffH, diffReport, err := h.getDiff(l, kubeCfg, nil, rel, h.state.plan.HelmDeployPlan.Namespace)
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
	client.WaitStrategy = helmkube.StatusWatcherStrategy
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
