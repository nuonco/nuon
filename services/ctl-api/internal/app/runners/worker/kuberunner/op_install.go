package runner

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/helm"
)

func (h *Activities) install(ctx context.Context, actionCfg *action.Configuration, req *InstallOrUpgradeRequest, prevRel *release.Release) (*release.Release, error) {
	l := zap.L()

	l.Info("loading chart")
	c, err := helm.GetChartByPath(h.config.OrgRunnerHelmChartDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart: %w", err)
	}
	releaseName := fmt.Sprintf("runner-%s", req.RunnerID)

	// get an install action "client"
	client := helm.DefaultInstall(actionCfg)
	// overrides some default values
	client.Devel = false
	// set values not provided by default install action "client" config
	client.CreateNamespace = true
	client.Namespace = req.Namespace
	client.ReleaseName = releaseName
	client.Timeout = req.Timeout
	client.DryRun = false

	if needsIntervention, err := helm.ReleaseNeedsManualIntervention(prevRel); needsIntervention {
		return nil, err
	}

	if helm.ReleaseNeedsCleanup(prevRel) {
		l.Info("cleaning up stuck release", zap.String("status", string(prevRel.Info.Status)))
		if err := helm.UninstallRelease(actionCfg, releaseName); err != nil {
			return nil, fmt.Errorf("unable to cleanup stuck release: %w", err)
		}
	} else if helm.ReleaseNeedsReplace(prevRel) {
		l.Info("replacing failed release", zap.String("status", string(prevRel.Info.Status)))
		client.Replace = true
	}

	l.Info("loading values")
	vals := h.getValues(req)
	mapVals, err := generics.ToMapstructure(vals)
	if err != nil {
		return nil, fmt.Errorf("unable to get mapstructure values: %w", err)
	}

	l.Info("running install")
	rel, err := client.RunWithContext(ctx, c, mapVals)
	if err != nil {
		return nil, fmt.Errorf("unable to install chart: %w", err)
	}

	return rel, nil
}
