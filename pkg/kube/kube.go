package kube

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetKubeConfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil && cfg != nil {
		return cfg, nil
	}
	home := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if cfg, err := clientcmd.BuildConfigFromFlags("", home); err == nil && cfg != nil {
		return cfg, nil
	}

	return nil, fmt.Errorf("failed to create k8s config")
}
