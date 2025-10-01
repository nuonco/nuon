package helm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	kube "helm.sh/helm/v4/pkg/kube"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/helm"
)

// HelmContentDiff represents a difference in the content of a Helm chart resource.
// It is used to capture changes between two states of a resource.
// Fields:
// - ApiPath: The API path of the resource (e.g., "apps/v1").
// - Name: The name of the resource.
// - Namespace: The namespace in which the resource resides.
// - Kind: The kind of the resource (e.g., "Deployment", "Service").
// - Before: The state of the resource before the change, typically serialized as a string.
// - After: The state of the resource after the change, typically serialized as a string.
type HelmContentDiff struct {
	ApiPath   string `json:"api,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Before    string `json:"before"`
	After     string `json:"after"`
}

func (hc *HelmContentDiff) parseRawName(s string) {
	// Parse format: "namespace, name, kind (apiBase)"
	if s == "" {
		return
	}

	// Use regex to parse the format: "namespace, name, kind (apiBase)"
	re := regexp.MustCompile(`^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 5 {
		hc.Namespace = strings.TrimSpace(matches[1])
		hc.Name = strings.TrimSpace(matches[2])
		hc.Kind = strings.TrimSpace(matches[3])
		hc.ApiPath = strings.TrimSpace(matches[4])
	}
}

type DeltaType int

const (
	// Common indicates the resource exists in both current and new state (no change)
	Common DeltaType = iota
	// DeltaTypeLeftOnly indicates the resource exists only in the current state (will be deleted)
	DeltaTypeLeftOnly = 1
	// DeltaTypeRightOnly indicates the resource exists only in the new state (will be added)
	DeltaTypeRightOnly = 2
)

// String returns a string representation for DeltaType.
func (t DeltaType) String() string {
	switch t {
	case Common:
		return " "
	case DeltaTypeLeftOnly:
		return "-"
	case DeltaTypeRightOnly:
		return "+"
	}
	return "?"
}

// HelmDiffcontentV2Entry represents a single diff entry within a Helm resource comparison.
// It contains the actual content difference and the type of change that occurred.
// Fields:
// - Payload: The actual content/text that represents the difference
// - Delta: The type of change (added, removed, or unchanged)
type HelmDiffcontentV2Entry struct {
	Payload string    `json:"payload"`
	Delta   DeltaType `json:"delta"`
}

// HelmDiffContentV2 represents the complete diff information for a single Kubernetes resource
// when comparing two Helm chart states. This is version 2 of the diff content structure.
// Fields:
// - Version: The version identifier for this diff format (always "2")
// - Name: The name of the Kubernetes resource being compared
// - Namespace: The namespace in which the resource resides
// - Kind: The kind of the Kubernetes resource (e.g., "Deployment", "Service", "ConfigMap")
// - ApiPath: The API path/version of the resource (e.g., "apps/v1", "v1")
// - Entries: A slice of individual diff entries that make up the complete difference for this resource
type HelmDiffContentV2 struct {
	Version string `json:"_version"`

	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	ApiPath   string `json:"api,omitempty"`

	Entries []HelmDiffcontentV2Entry `json:"entries"`
}

func (hc *HelmDiffContentV2) parseRawName(s string) {
	// Parse format: "namespace, name, kind (apiBase)"
	if s == "" {
		return
	}

	// Use regex to parse the format: "namespace, name, kind (apiBase)"
	re := regexp.MustCompile(`^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 5 {
		hc.Namespace = strings.TrimSpace(matches[1])
		hc.Name = strings.TrimSpace(matches[2])
		hc.Kind = strings.TrimSpace(matches[3])
		hc.ApiPath = strings.TrimSpace(matches[4])
	}
}

func newHelmContentV2() HelmDiffContentV2 {
	return HelmDiffContentV2{
		Version: "2",
	}
}

func (h *handler) upgrade_diff(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config) (string, *[]HelmDiffContentV2, error) {
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
	client.Atomic = false
	client.SkipCRDs = false
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Description = ""
	client.ResetValues = false
	client.ReuseValues = false
	client.Recreate = false
	client.MaxHistory = 0
	client.CleanupOnFail = false
	client.Force = false

	l.Info("calculating helm diff")
	rel, err := client.RunWithContext(ctx, prevRel.Name, chart, values)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}
	l.Info("parsing previous and current release manifests")
	diff, diffReport, err := h.getDiff(l, kubeCfg, prevRel, rel, h.state.plan.HelmDeployPlan.Namespace)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}

	h.state.outputs = map[string]interface{}{"diff": diff}

	contentDiff := make([]HelmDiffContentV2, 0)
	for _, re := range diffReport.Entries {
		d := newHelmContentV2()
		d.parseRawName(re.Key)

		for _, diff := range re.Diffs {
			d.Entries = append(d.Entries, HelmDiffcontentV2Entry{
				Payload: diff.Payload,
				Delta:   DeltaType(diff.Delta),
			})
		}

		contentDiff = append(contentDiff, d)
	}

	return string(diff), &contentDiff, nil
}

func (h *handler) installDiff(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config) (string, *[]HelmDiffContentV2, error) {
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
	client.Atomic = false
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
	diff, diffReport, err := h.getDiff(l, kubeCfg, nil, rel, h.state.plan.HelmDeployPlan.Namespace)

	h.state.outputs = map[string]interface{}{"diff": diff}

	contentDiff := make([]HelmDiffContentV2, 0)
	for _, re := range diffReport.Entries {
		d := newHelmContentV2()
		d.parseRawName(re.Key)

		for _, diff := range re.Diffs {
			d.Entries = append(d.Entries, HelmDiffcontentV2Entry{
				Payload: diff.Payload,
				Delta:   DeltaType(diff.Delta),
			})
		}

		contentDiff = append(contentDiff, d)
	}

	return string(diff), &contentDiff, nil
}

func (h *handler) uninstallDiff(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, kubeCfg *rest.Config, prevRel *release.Release) (string, *[]HelmDiffContentV2, error) {
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

	diff, diffReport, err := h.getDiff(l, kubeCfg, prevRel, nil, h.state.plan.HelmDeployPlan.Namespace)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to execute with dry-run")
	}

	h.state.outputs = map[string]interface{}{"diff": diff}

	contentDiff := make([]HelmDiffContentV2, 0)
	for _, re := range diffReport.Entries {
		d := newHelmContentV2()
		d.parseRawName(re.Key)

		for _, diff := range re.Diffs {
			d.Entries = append(d.Entries, HelmDiffcontentV2Entry{
				Payload: diff.Payload,
				Delta:   DeltaType(diff.Delta),
			})
		}

		contentDiff = append(contentDiff, d)
	}

	return string(diff), &contentDiff, nil
}
