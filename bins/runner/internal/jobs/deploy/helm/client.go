package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) actionInit(ctx context.Context, l *zap.Logger) (*action.Configuration, *rest.Config, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.plan.HelmDeployPlan.ClusterInfo)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get kube config")
	}

	helmCfg, err := helm.Client(l, kubeCfg, h.state.plan.HelmDeployPlan.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmCfg, kubeCfg, nil
}

func (h *handler) getHelmReleaseStore(ctx context.Context, kubeCfg *rest.Config) (*storage.Storage, error) {
	k8sClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, err
	}

	var store *storage.Storage
	switch h.state.plan.HelmDeployPlan.StorageDriver {
	case "configmap", "configmaps":
		// corev1.ConfigMapInterface
		configmaps := k8sClient.CoreV1().ConfigMaps(h.state.plan.HelmDeployPlan.Namespace)
		d := driver.NewConfigMaps(configmaps)
		store = storage.Init(d)
	case "secrets":
		// corev1.SecretsInterface
		secrets := k8sClient.CoreV1().Secrets(h.state.plan.HelmDeployPlan.Namespace)
		d := driver.NewSecrets(secrets)
		store = storage.Init(d)
	default:
		return nil, errors.New("unsupported driver type " + h.state.plan.HelmDeployPlan.StorageDriver)
	}
	return store, nil
}
