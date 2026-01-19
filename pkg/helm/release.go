package helm

import (
	"fmt"
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

func IsReleaseDeployed(rel *release.Release) bool {
	if rel == nil {
		return false
	}
	return rel.Info.Status == release.StatusDeployed
}

func ReleaseNeedsReplace(rel *release.Release) bool {
	if rel == nil {
		return false
	}
	return rel.Info.Status == release.StatusFailed
}

func ReleaseNeedsCleanup(rel *release.Release) bool {
	if rel == nil {
		return false
	}
	return rel.Info.Status == release.StatusPendingInstall
}

func ReleaseNeedsManualIntervention(rel *release.Release) (bool, error) {
	if rel == nil {
		return false, nil
	}
	status := rel.Info.Status
	if status == release.StatusPendingUpgrade {
		return true, fmt.Errorf("release is stuck in '%s' state from a previous failed upgrade. Run 'helm rollback %s' to recover", status, rel.Name)
	}
	if status == release.StatusPendingRollback {
		return true, fmt.Errorf("release is stuck in '%s' state from a previous failed rollback. Run 'helm uninstall %s' and redeploy to recover", status, rel.Name)
	}
	return false, nil
}

func UninstallRelease(cfg *action.Configuration, name string) error {
	client := action.NewUninstall(cfg)
	client.IgnoreNotFound = true
	_, err := client.Run(name)
	return err
}
