package helm

import (
	"strings"

	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
)

func GetRelease(cfg *action.Configuration, name string) (*release.Release, error) {
	res, err := action.NewGet(cfg).Run(name)
	if err != nil {
		if strings.Contains(err.Error(), "release: not found") {
			return nil, nil
		}

		return nil, err
	}

	return res, nil
}
