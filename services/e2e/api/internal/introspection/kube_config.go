package introspection

import (
	"context"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func (s *svc) getKubeConfig(ctx context.Context) (*rest.Config, error) {
	home := filepath.Join(homedir.HomeDir(), ".kube", "config")
	localKubeCfg, err := clientcmd.BuildConfigFromFlags("", home)
	if err == nil {
		return localKubeCfg, nil
	}

	kubeCfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get in cluster config: %w", err)
	}

	return kubeCfg, nil
}
