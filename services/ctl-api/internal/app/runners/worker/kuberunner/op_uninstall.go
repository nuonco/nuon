package runner

import (
	"context"
	"fmt"

	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"

	"github.com/nuonco/nuon/pkg/helm"
)

func (a *Activities) uninstall(ctx context.Context, actionCfg *action.Configuration, runnerID string) error {
	releaseName := fmt.Sprintf("runner-%s", runnerID)
	prevRel, err := helm.GetRelease(actionCfg, releaseName)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	if prevRel == nil {
		return nil
	}

	client := action.NewUninstall(actionCfg)
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.Timeout = defaultHelmOperationTimeout

	_, err = client.Run(prevRel.Name)
	if err != nil {
		return fmt.Errorf("unable to uninstall previous release: %w", err)
	}

	return nil
}
