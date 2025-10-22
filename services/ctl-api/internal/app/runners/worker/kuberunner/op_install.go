package runner

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *Activities) install(ctx context.Context, actionCfg *action.Configuration, req *InstallOrUpgradeRequest) (*release.Release, error) {
	l := zap.L()

	l.Info("loading chart")
	c, err := helm.GetChartByPath(h.config.OrgRunnerHelmChartDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart: %w", err)
	}
	// get an install action "client"
	client := helm.DefaultInstall(actionCfg)
	// overrides some default values
	client.Devel = false
	// set values not provided by default install action "client" config
	client.CreateNamespace = true
	client.Namespace = req.Namespace
	client.ReleaseName = fmt.Sprintf("runner-%s", req.RunnerID)
	client.Timeout = req.Timeout
	client.DryRun = false

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
