package helm

import (
	"strings"

	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// releaseNotFound is the SDK's error text for a release that has never been
// stored. It is matched on the string because the driver error is wrapped
// several layers deep by the time it surfaces from an action.
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

// IsPending reports whether a release is parked mid-operation. Helm writes one
// of these statuses before it starts touching the cluster and only overwrites it
// once the operation finishes, so a release left in one of them is a rollout
// whose driver went away — a crashed or cancelled runner, or a job that timed
// out. Helm then refuses every subsequent operation on the release, and no
// amount of retrying clears it.
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

// LastGoodRevision returns the highest revision that finished a rollout, which
// is the only safe target to roll a stuck release back to.
//
// "Finished" means deployed or superseded. A failed revision is excluded: it ran
// and did not work, so returning to it would trade one broken release for
// another. Pending revisions are excluded for the same reason they are the
// problem. The bool is false when no such revision exists, which is the normal
// case for a release stuck on its very first install — there is nothing behind
// it to go back to.
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
