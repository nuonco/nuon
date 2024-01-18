package runner

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) installEKSRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) error {
	l := workflow.GetLogger(ctx)

	// install waypoint
	wpChart, err := helm.LoadChart(w.cfg.WaypointChartDir)
	if err != nil {
		return fmt.Errorf("unable to load helm chart: %w", err)
	}

	chart := &helm.Chart{
		Name:    wpChart.Metadata.Name + "-runner",
		Version: wpChart.Metadata.Version,
		Dir:     w.cfg.WaypointChartDir,
	}

	orgServerAddr := client.DefaultOrgServerAddress(req.OrgId, w.cfg.OrgServerRootDomain)

	// get waypoint server cookie
	gwscReq := GetWaypointServerCookieRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		ClusterInfo:          w.clusterInfo,
	}
	gwscResp, err := w.getWaypointServerCookie(ctx, gwscReq)
	if err != nil {
		err = fmt.Errorf("failed to get waypoint server cookie: %w", err)
		return err
	}

	l.Info("runner chart version", wpChart.Metadata.Version)
	iwReq := InstallWaypointRequest{
		InstallID:       req.InstallId,
		Namespace:       req.InstallId,
		ReleaseName:     fmt.Sprintf("wp-%s", req.InstallId),
		Chart:           chart,
		Atomic:          false,
		CreateNamespace: true,
		ClusterInfo: kube.ClusterInfo{
			ID:             req.EksClusterInfo.Id,
			Endpoint:       req.EksClusterInfo.Endpoint,
			CAData:         req.EksClusterInfo.CaData,
			TrustedRoleARN: req.EksClusterInfo.TrustedRoleArn,
		},

		RunnerConfig: RunnerConfig{
			OdrIAMRoleArn: req.OdrIamRoleArn,
			Cookie:        gwscResp.Cookie,
			ID:            req.InstallId,
			ServerAddr:    orgServerAddr,
		},
	}
	_, err = w.installWaypoint(ctx, iwReq)
	if err != nil {
		err = fmt.Errorf("failed to install waypoint: %w", err)
		return err
	}

	// TODO(jm): this is now fixed in the helm chart, so this can be removed barring additional testing.
	crbReq := CreateRoleBindingRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		InstallID:            req.InstallId,
		NamespaceName:        req.InstallId,
		ClusterInfo: kube.ClusterInfo{
			ID:             req.EksClusterInfo.Id,
			Endpoint:       req.EksClusterInfo.Endpoint,
			CAData:         req.EksClusterInfo.CaData,
			TrustedRoleARN: req.EksClusterInfo.TrustedRoleArn,
		},
	}
	_, err = w.createRoleBinding(ctx, crbReq)
	if err != nil {
		err = fmt.Errorf("failed to create role_binding for runner: %w", err)
		return err
	}

	awrReq := AdoptWaypointRunnerRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	_, err = w.adoptWaypointRunner(ctx, awrReq)
	if err != nil {
		err = fmt.Errorf("failed to adopt waypoint runner: %w", err)
		return err
	}

	cwrpReq := CreateWaypointRunnerProfileRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		InstallID:            req.InstallId,
		OrgID:                req.OrgId,
		AwsRegion:            req.Region,
		ClusterInfo:          w.clusterInfo,
	}
	_, err = w.createWaypointRunnerProfile(ctx, cwrpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint runner profile: %w", err)
		return err
	}

	return nil
}
