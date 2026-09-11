package helm

import (
	"fmt"

	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// ConfirmUninstalled resolves a failed uninstall against the release store and
// returns nil only when the release is genuinely gone.
//
// Helm reports "release: not found" both when there was nothing to remove and
// when a real uninstall got partway through and then failed to purge the release
// record — a 404 from the storage backend on the record it had just served. The
// first is success, the second is not, and the error alone cannot tell them
// apart. Re-reading the release can: if it is still stored, the uninstall did not
// finish. Anything that is not a not-found error is returned untouched.
func ConfirmUninstalled(cfg *action.Configuration, name string, uninstallErr error) error {
	if uninstallErr == nil {
		return nil
	}
	if !IsReleaseNotFound(uninstallErr) {
		return uninstallErr
	}

	rel, err := GetRelease(cfg, name)
	if err != nil {
		return fmt.Errorf(
			"uninstall of release %s reported it was not found and re-reading it failed, so it cannot be confirmed removed: %w (uninstall error: %v)",
			name, err, uninstallErr,
		)
	}
	if rel != nil {
		return fmt.Errorf(
			"uninstall of release %s reported it was not found, but it is still stored at revision %d with status %s: %w",
			name, rel.Version, releaseStatus(rel), uninstallErr,
		)
	}

	return nil
}

func releaseStatus(rel *release.Release) string {
	if rel == nil || rel.Info == nil {
		return "unknown"
	}

	return string(rel.Info.Status)
}
