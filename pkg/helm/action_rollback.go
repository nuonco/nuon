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

	// Empty is rejected by the SDK; "auto" reuses the target revision's method.
	client.ServerSideApply = "auto"

	// wait
	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	return client
}

// Rollback returns the release to revision, which must still exist in history.
// Needs no chart on disk: the SDK rebuilds it from the stored revision.
func Rollback(actionCfg *action.Configuration, name string, revision int, timeout time.Duration) error {
	client := DefaultRollback(actionCfg)
	client.Version = revision
	client.Timeout = timeout

	if err := client.Run(name); err != nil {
		return fmt.Errorf("unable to roll back release %s to revision %d: %w", name, revision, err)
	}

	return nil
}
