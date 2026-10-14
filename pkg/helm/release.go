package helm

import (
	"errors"
	"fmt"
	"strings"

	"helm.sh/helm/v4/pkg/action"
	relapi "helm.sh/helm/v4/pkg/release"
	rcommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
)

// AsRelease unwraps a Helm 4.1 action result into the v1 release the rest of
// the runner reads. Helm returns the same value as an empty interface.
func AsRelease(rel relapi.Releaser) (*release.Release, error) {
	switch r := rel.(type) {
	case *release.Release:
		return r, nil
	case release.Release:
		return &r, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported helm release type %T", rel)
	}
}

func asReleases(rels []relapi.Releaser) ([]*release.Release, error) {
	out := make([]*release.Release, 0, len(rels))
	for _, item := range rels {
		rel, err := AsRelease(item)
		if err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	return out, nil
}

const releaseNotFound = "release: not found"

// IsReleaseNotFound reports whether err means no such release is stored.
//
// Both forms have to be handled. Helm's own errors keep the sentinel in the
// chain — including through the joined errors a partially failed uninstall
// returns — but several paths format it with %s or %v instead of %w, leaving
// only the message behind.
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

	return AsRelease(res)
}

// History returns every stored revision of a release, oldest first. A release
// that was never stored yields an empty history rather than an error, matching
// GetRelease's treatment of the same condition.
func History(cfg *action.Configuration, name string) ([]*release.Release, error) {
	res, err := action.NewHistory(cfg).Run(name)
	if err != nil {
		if IsReleaseNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return asReleases(res)
}

// IsPending reports whether a release is parked mid-operation, which helm
// refuses every further operation on until it is recovered.
func IsPending(rel *release.Release) bool {
	if rel == nil || rel.Info == nil {
		return false
	}
	switch rel.Info.Status {
	case rcommon.StatusPendingInstall,
		rcommon.StatusPendingUpgrade,
		rcommon.StatusPendingRollback:
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
		case rcommon.StatusDeployed, rcommon.StatusSuperseded:
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
	case rcommon.StatusDeployed,
		rcommon.StatusFailed,
		rcommon.StatusSuperseded:
		return true
	default:
		return false
	}
}
