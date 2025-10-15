package runner

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *Activities) upgrade(ctx context.Context, actionCfg *action.Configuration, req *InstallOrUpgradeRequest) (*release.Release, error) {
	l := zap.L()
	l.Info("loading chart")
	c, err := helm.GetChartByPath(h.config.OrgRunnerHelmChartDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart: %w", err)
	}

	l.Info("fetching previous release")
	releaseName := fmt.Sprintf("runner-%s", req.RunnerID)
	prevRel, err := helm.GetRelease(actionCfg, releaseName)
	if err != nil {
		return nil, fmt.Errorf("unable to get previous helm release: %w", err)
	}

	// We have a previous release, upgrade.
	client := action.NewUpgrade(actionCfg)
	client.DryRun = false
	client.DisableHooks = false
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.WaitForJobs = false
	client.Devel = false
	client.DependencyUpdate = true
	client.Timeout = req.Timeout
	client.Namespace = req.Namespace
	client.SkipCRDs = false
	client.SubNotes = true
	client.DisableOpenAPIValidation = false
	client.Description = ""
	client.ResetValues = false
	client.ReuseValues = false
	client.MaxHistory = 0
	client.CleanupOnFail = false

	l.Info("loading values")
	vals := h.getValues(req)
	mapVals, err := generics.ToMapstructure(vals)
	if err != nil {
		return nil, fmt.Errorf("unable to get mapstructure values: %w", err)
	}

	l.Info("upgrading helm release")
	rel, err := client.RunWithContext(ctx, prevRel.Name, c, mapVals)
	if err != nil {
		return nil, fmt.Errorf("unable to upgrade helm release: %w", err)
	}

	return rel, nil
}
