package helm

import (
	"errors"
	"strings"

	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
)

const releaseNotFound = "release: not found"

func IsReleaseNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, driver.ErrReleaseNotFound) {
		return true
	}

	return strings.Contains(err.Error(), releaseNotFound)
}

func GetRelease(cfg *action.Configuration, name string) (*release.Release, error) {
	res, err := action.NewGet(cfg).Run(name)
	if err != nil {
		if IsReleaseNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return res, nil
}

func History(cfg *action.Configuration, name string) ([]*release.Release, error) {
	res, err := action.NewHistory(cfg).Run(name)
	if err != nil {
		if IsReleaseNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return res, nil
}

func IsPending(rel *release.Release) bool {
	if rel == nil || rel.Info == nil {
		return false
	}
	switch rel.Info.Status {
	case release.StatusPendingInstall,
		release.StatusPendingUpgrade,
		release.StatusPendingRollback:
		return true
	default:
		return false
	}
}

func LastGoodRevision(history []*release.Release) (int, bool) {
	best := 0
	for _, rel := range history {
		if rel == nil || rel.Info == nil {
			continue
		}
		switch rel.Info.Status {
		case release.StatusDeployed, release.StatusSuperseded:
			if rel.Version > best {
				best = rel.Version
			}
		}
	}

	return best, best > 0
}

func ShouldUpgrade(rel *release.Release) bool {
	if rel == nil {
		return false
	}
	switch rel.Info.Status {
	case release.StatusDeployed,
		release.StatusFailed,
		release.StatusSuperseded:
		return true
	default:
		return false
	}
}
