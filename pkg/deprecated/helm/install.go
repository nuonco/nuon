package helm

import (
	"context"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/mono/pkg/kube"
)

type chartLoader interface {
	load(string, action.ChartPathOptions) (*chart.Chart, error)
}

type helmInstallRunner struct {
	chartLoader
}

func isHasNoDeployedReleasesError(err error) bool {
	return strings.Contains(err.Error(), "has no deployed releases")
}

func isIgnoredInstallError(err error) bool {
	return strings.Contains(err.Error(), "manifests contain a resource that already exists") ||
		strings.Contains(err.Error(), "cannot re-use a name that is still in use")
}

type chrtLoader struct{}

var _ chartLoader = (*chrtLoader)(nil)

func (l *chrtLoader) load(name string, client action.ChartPathOptions) (*chart.Chart, error) {
	path, err := client.LocateChart(name, cli.New())
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	return loader.Load(path)
}

type Chart struct {
	Name    string `json:"name" validate:"required"`
	URL     string `json:"url"`
	Dir     string `json:"dir"`
	Version string `json:"version" validate:"required"`
}

type InstallConfig struct {
	Namespace   string
	ReleaseName string
	Chart       *Chart
	Atomic      bool
	Values      map[string]interface{} `faker:"-"`
	Logger      Logger                 `faker:"-"`

	// These are exposed for testing. Do not use otherwise
	CreateNamespace bool
	Kubeconfig      *rest.Config `faker:"-"`
}

type installer struct{}

func NewInstaller() *installer {
	return &installer{}
}

func (i *installer) Install(ctx context.Context, config *InstallConfig) (*release.Release, error) {
	l := config.Logger
	if l == nil {
		l = &fmtLogger{}
	}

	var err error
	kCfg := config.Kubeconfig
	if kCfg == nil {
		kCfg, err = kube.GetKubeConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get kube config: %w", err)
		}
		kCfg.Burst = 1000
	}

	clientset, err := kubernetes.NewForConfig(kCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}

	actionConfig := new(action.Configuration)
	upgr := action.NewUpgrade(actionConfig)
	upgr.Atomic = config.Atomic
	upgr.Namespace = config.Namespace
	upgr.RepoURL = config.Chart.URL
	upgr.Version = config.Chart.Version
	upgr.ResetValues = true

	rcg := &restClientGetter{RestConfig: kCfg, Clientset: clientset}
	actionLogger := func(format string, v ...interface{}) { l.Debug(fmt.Sprintf(format, v...)) }

	err = actionConfig.Init(rcg, upgr.Namespace, helmDriver, actionLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm config: %w", err)
	}

	inputName := config.Chart.Name
	if config.Chart.Dir != "" {
		inputName = config.Chart.Dir
	}

	// locate the chart
	path, err := upgr.LocateChart(inputName, cli.New())
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	chrt, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	rel, err := upgr.RunWithContext(ctx, config.ReleaseName, chrt, config.Values)
	if err != nil && isHasNoDeployedReleasesError(err) {
		inst := action.NewInstall(actionConfig)

		inst.CreateNamespace = true
		inst.Atomic = config.Atomic
		inst.Namespace = config.Namespace
		inst.RepoURL = config.Chart.URL
		inst.Version = config.Chart.Version
		inst.ReleaseName = config.ReleaseName
		inst.Replace = true
		return inst.RunWithContext(ctx, chrt, config.Values)
	}
	if err != nil && !isIgnoredInstallError(err) {
		return nil, fmt.Errorf("failed to upgrade chart: %w", err)
	}
	return rel, nil
}
