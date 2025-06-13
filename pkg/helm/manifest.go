package helm

import "helm.sh/helm/v4/pkg/action"

func GetManifest(cfg *action.Configuration, name string) (any, error) {
	action.NewGet(cfg)
	return nil, nil
}
