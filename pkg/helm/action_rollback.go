package helm

import (
	"fmt"
	"time"

	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"
)

func DefaultRollback(actionCfg *action.Configuration) *action.Rollback {
	client := action.NewRollback(actionCfg)
	return ConfigureDefaultRollback(client)
}

func ConfigureDefaultRollback(client *action.Rollback) *action.Rollback {
	client.CleanupOnFail = false
	client.DisableHooks = false
	client.DryRun = false
	client.ForceReplace = false
	client.MaxHistory = 0

	// "auto" reuses whatever apply method the target revision was rolled out
	// with. An empty value is rejected outright by the SDK, and hardcoding
	// "false" would silently change the apply method of a release that was
	// created server-side.
	client.ServerSideApply = "auto"

	// wait
	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	return client
}

// Rollback returns the release to an earlier revision. revision must name a
// revision that still exists in the release history; the SDK does not accept a
// relative offset and rejects 0 as "the previous one" only by convention, which
// is too implicit to rely on for a recovery path.
//
// Unlike install and upgrade this needs no chart on disk: the SDK rebuilds the
// target release from the chart and values stored in the revision itself.
func Rollback(actionCfg *action.Configuration, name string, revision int, timeout time.Duration) error {
	client := DefaultRollback(actionCfg)
	client.Version = revision
	client.Timeout = timeout

	if err := client.Run(name); err != nil {
		return fmt.Errorf("unable to roll back release %s to revision %d: %w", name, revision, err)
	}

	return nil
}
