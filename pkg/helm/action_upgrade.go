package helm

import (
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"
)

func DefaultUpgrade(actionCfg *action.Configuration) *action.Upgrade {
	client := action.NewUpgrade(actionCfg)
	return ConfigureDefaultUpgrade(client)
}

func ConfigureDefaultUpgrade(client *action.Upgrade) *action.Upgrade {
	client.CleanupOnFail = false
	client.DependencyUpdate = true
	client.Description = ""

	client.DisableHooks = false
	client.DisableOpenAPIValidation = false
	client.MaxHistory = 0
	client.ResetValues = false
	client.ReuseValues = false
	client.SkipCRDs = false
	client.SubNotes = true
	// wait
	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	return client

}
