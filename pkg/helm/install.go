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

type installRunner interface {
	install(context.Context, *action.Install, string, map[string]interface{}) (*release.Release, error)
}

type chartLoader interface {
	load(string, action.ChartPathOptions) (*chart.Chart, error)
}

type helmInstallRunner struct {
	chartLoader
}

var _ installRunner = (*helmInstallRunner)(nil)

func (h *helmInstallRunner) install(ctx context.Context, client *action.Install, chartName string, values map[string]interface{}) (*release.Release, error) {
	chrt, err := h.load(chartName, client.ChartPathOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	rel, err := client.RunWithContext(ctx, chrt, values)
	if err != nil && !isIgnoredInstallError(err) {
		return nil, fmt.Errorf("failed to install chart: %w", err)
	}
	return rel, nil
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
	URL     string `json:"url" validate:"required"`
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
	Kubeconfig      *rest.Config  `faker:"-"`
	installer       installRunner `faker:"-"`
}

type installer struct{}

func NewInstaller() *installer {
	return &installer{}
}

func (i *installer) Install(ctx context.Context, config *InstallConfig) (*release.Release, error) {
	if config.installer == nil {
		config.installer = &helmInstallRunner{&chrtLoader{}}
	}

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
	client := action.NewInstall(actionConfig)

	// TODO(jdt): set timeout based on activity timeout?
	// client.Timeout = ?
	client.Atomic = config.Atomic
	client.Namespace = config.Namespace
	client.RepoURL = config.Chart.URL
	client.Version = config.Chart.Version
	client.ReleaseName = config.ReleaseName
	client.CreateNamespace = config.CreateNamespace
	client.Replace = true

	rcg := &restClientGetter{RestConfig: kCfg, Clientset: clientset}
	actionLogger := func(format string, v ...interface{}) { l.Debug(fmt.Sprintf(format, v...)) }

	err = actionConfig.Init(rcg, client.Namespace, helmDriver, actionLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm config: %w", err)
	}

	r, err := config.installer.install(ctx, client, config.Chart.Name, config.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to install release: %w", err)
	}
	return r, nil
}
