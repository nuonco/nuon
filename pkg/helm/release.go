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

// ShouldUpgrade returns true when a release exists in a state that warrants
// an upgrade rather than a fresh install. This includes deployed releases as
// well as failed releases — Helm's upgrade action natively handles upgrading
// over a failed release.
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
