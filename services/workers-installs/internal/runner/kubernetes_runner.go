package runner

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) installKubernetesRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) error {
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

	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId)

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

	clusterInfo := kube.ClusterInfo{}
	if req.EksClusterInfo != nil {
		clusterInfo.ID = req.EksClusterInfo.Id
		clusterInfo.Endpoint = req.EksClusterInfo.Endpoint
		clusterInfo.CAData = req.EksClusterInfo.CaData
		clusterInfo.TrustedRoleARN = req.EksClusterInfo.TrustedRoleArn
	}
	if req.AksClusterInfo != nil {
		clusterInfo.KubeConfig = req.AksClusterInfo.KubeConfig
	}

	l.Info("runner chart version", wpChart.Metadata.Version)
	iwReq := InstallWaypointRequest{
		InstallID:       req.InstallId,
		Namespace:       req.InstallId,
		ReleaseName:     fmt.Sprintf("wp-%s", req.InstallId),
		Chart:           chart,
		Atomic:          false,
		CreateNamespace: true,
		ClusterInfo:     clusterInfo,
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
		ClusterInfo:          clusterInfo,
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
		RunnerType:           installsv1.RunnerType_RUNNER_TYPE_AWS_EKS,
	}
	_, err = w.createWaypointRunnerProfile(ctx, cwrpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint runner profile: %w", err)
		return err
	}

	return nil
}

func (w *wkflow) uninstallKubernetesRunner(ctx workflow.Context, req *runnerv1.DeprovisionRunnerRequest) error {
	l := workflow.GetLogger(ctx)

	// NOTE(jm): this is not a long term solution, eventually we will manage both the runner and the different
	// components using nuon components, and then will just remove these by orchestrating the executors upstream.
	//
	// howe  er, for now, until this all works we just "cheat" and delete the builtin namespace
	listResp, err := w.execListNamespaces(ctx, ListNamespacesRequest{
		AppID:     req.AppId,
		OrgID:     req.OrgId,
		InstallID: req.InstallId,
	})
	if err != nil {
		l.Debug("unable to list namespaces", "error", err)
		return fmt.Errorf("unable to delete namespace: %w", err)
	}

	for _, namespace := range listResp.Namespaces {
		if generics.SliceContains(namespace, terraformManagedNamespaces) {
			continue
		}

		_, err = w.execDeleteNamespace(ctx, DeleteNamespaceRequest{
			AppID:     req.AppId,
			OrgID:     req.OrgId,
			InstallID: req.InstallId,
			Namespace: namespace,
		})
		if err != nil {
			l.Debug("unable to delete namespace activity", "error", err)
			return fmt.Errorf("unable to delete namespace: %w", err)
		}
	}
	return nil
}
