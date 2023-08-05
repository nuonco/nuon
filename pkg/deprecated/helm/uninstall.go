package helm

import (
	"context"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type uninstallRunner interface {
	uninstall(context.Context, *action.Uninstall, string) (*release.UninstallReleaseResponse, error)
}
type helmUninstaller struct{}

var _ uninstallRunner = (*helmUninstaller)(nil)

func (w *helmUninstaller) uninstall(ctx context.Context, client *action.Uninstall, releaseName string) (*release.UninstallReleaseResponse, error) {
	resp, err := client.Run(releaseName)
	if err != nil && !helmUninstallIgnore(err) {
		return nil, fmt.Errorf("failed to uninstall chart: %w", err)
	}

	return resp, nil
}

type UninstallConfig struct {
	Namespace   string
	ReleaseName string
	Logger      Logger       `faker:"-"`
	Kubeconfig  *rest.Config `faker:"-"`

	uninstaller uninstallRunner `faker:"-"`
}

func helmUninstallIgnore(err error) bool {
	s := err.Error()
	return strings.Contains(s, "Release not loaded")
}

type uninstaller struct{}

func NewUninstaller() *uninstaller {
	return &uninstaller{}
}

func (u *uninstaller) Uninstall(ctx context.Context, cfg *UninstallConfig) (*release.UninstallReleaseResponse, error) {
	if cfg.uninstaller == nil {
		cfg.uninstaller = &helmUninstaller{}
	}

	l := cfg.Logger
	if l == nil {
		l = &fmtLogger{}
	}

	clientset, err := kubernetes.NewForConfig(cfg.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create kube client: %w", err)
	}
	rcg := &RestClientGetter{RestConfig: cfg.Kubeconfig, Clientset: clientset}
	actionLogger := func(format string, v ...interface{}) { l.Debug(format, v) }

	actionConfig := new(action.Configuration)
	client := action.NewUninstall(actionConfig)

	err = actionConfig.Init(rcg, cfg.Namespace, helmDriver, actionLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm config: %w", err)
	}

	resp, err := cfg.uninstaller.uninstall(ctx, client, cfg.ReleaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to uninstall release: %w", err)
	}

	return resp, nil
}
