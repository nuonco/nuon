package helm

import (
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"
)

func DefaultInstall(actionCfg *action.Configuration) *action.Install {
	client := action.NewInstall(actionCfg)
	return ConfigureDefaultInstall(client)
}

// useful in case we want to configure a client that has already been created
// NOTE(fd): these default values were yoinked from the runner install code
func ConfigureDefaultInstall(client *action.Install) *action.Install {
	client.ClientOnly = false
	client.DisableHooks = false

	client.DependencyUpdate = true
	client.Description = ""
	client.Devel = true
	client.DisableOpenAPIValidation = false
	client.GenerateName = false
	client.NameTemplate = ""
	client.OutputDir = ""
	client.Replace = false
	client.SkipCRDs = false
	client.SubNotes = true

	// wait strategy
	client.WaitForJobs = false
	client.WaitStrategy = kube.StatusWatcherStrategy

	return client
}
