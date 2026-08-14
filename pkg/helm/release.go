package helm

import (
	"strings"

	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// Matched on the string: the driver error is wrapped several layers deep.
const releaseNotFound = "release: not found"

func GetRelease(cfg *action.Configuration, name string) (*release.Release, error) {
	res, err := action.NewGet(cfg).Run(name)
	if err != nil {
		if strings.Contains(err.Error(), releaseNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return res, nil
}

// History returns every stored revision of a release, oldest first. A release
// that was never stored yields an empty history rather than an error, matching
// GetRelease's treatment of the same condition.
func History(cfg *action.Configuration, name string) ([]*release.Release, error) {
	res, err := action.NewHistory(cfg).Run(name)
	if err != nil {
		if strings.Contains(err.Error(), releaseNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return res, nil
}

// IsPending reports whether a release is parked mid-operation, which helm
// refuses every further operation on until it is recovered.
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

// LastGoodRevision returns the highest revision that finished a rollout. Failed
// revisions are excluded: returning to one trades a stuck release for a broken
// one. False means nothing to roll back to, as on a first install.
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
