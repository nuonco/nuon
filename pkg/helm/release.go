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

func GetReleaseHistory(cfg *action.Configuration, name string, max int) ([]*release.Release, error) {
	history := action.NewHistory(cfg)
	history.Max = max

	releases, err := history.Run(name)
	if err != nil {
		if strings.Contains(err.Error(), "release: not found") {
			return nil, nil
		}

		return nil, err
	}

	return releases, nil
}

func FindLastDeployedVersion(cfg *action.Configuration, name string) (int, error) {
	releases, err := GetReleaseHistory(cfg, name, 10)
	if err != nil {
		return 0, err
	}

	if releases == nil {
		return 0, nil
	}

	// Find the most recent release with deployed status
	for i := len(releases) - 1; i >= 0; i-- {
		rel := releases[i]
		if rel.Info.Status == release.StatusDeployed {
			return rel.Version, nil
		}
	}

	return 0, nil
}
