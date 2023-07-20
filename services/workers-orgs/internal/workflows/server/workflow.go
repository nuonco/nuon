package server

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/helm/waypoint"
	"github.com/powertoolsdev/mono/pkg/kube"
	serverv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/server/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultWaypointServerPort int32 = 9701
)

type BootstrapError struct{}

func (b BootstrapError) Error() string {
	return "client already bootstrapped"
}

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

func (w wkflow) ProvisionServer(ctx workflow.Context, req *serverv1.ProvisionServerRequest) (*serverv1.ProvisionServerResponse, error) {
	resp := &serverv1.ProvisionServerResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	serverDomain := client.DefaultOrgServerDomain(w.cfg.WaypointServerRootDomain, req.OrgId)
	waypointServerAddr := client.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId)
	clusterInfo := kube.ClusterInfo{
		ID:             w.cfg.OrgsK8sClusterID,
		Endpoint:       w.cfg.OrgsK8sPublicEndpoint,
		CAData:         w.cfg.OrgsK8sCAData,
		TrustedRoleARN: w.cfg.OrgsK8sRoleArn,
	}

	act := NewActivities(nil)

	cnReq := CreateNamespaceRequest{
		NamespaceName: req.OrgId,
		ClusterInfo:   clusterInfo,
	}
	_, err := createNamespace(ctx, act, cnReq)
	if err != nil {
		return resp, fmt.Errorf("failed to create namespace: %w", err)
	}

	l.Debug("installing waypoint server")
	chart := &helm.Chart{
		Name:    waypoint.DefaultChart.Name,
		Version: waypoint.DefaultChart.Version,
		Dir:     w.cfg.WaypointChartDir,
	}
	_, err = installWaypointServer(ctx, act, InstallWaypointServerRequest{
		Namespace:   req.OrgId,
		ReleaseName: fmt.Sprintf("wp-%s", req.OrgId),
		Domain:      serverDomain,
		OrgID:       req.OrgId,
		Chart:       chart,
		Atomic:      false,
		ClusterInfo: clusterInfo,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to install waypoint: %w", err)
	}

	l.Debug("pinging waypoint server until it responds")
	_, err = pingWaypointServer(ctx, act, PingWaypointServerRequest{
		Addr:    waypointServerAddr,
		Timeout: time.Minute * 10,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to ping waypoint: %w", err)
	}

	l.Debug("bootstrapping waypoint server")
	_, err = bootstrapWaypointServer(ctx, act, BootstrapWaypointServerRequest{
		ServerAddr:     waypointServerAddr,
		TokenNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgID:          req.OrgId,
		ClusterInfo:    clusterInfo,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to bootstrap waypoint: %w", err)
	}

	l.Debug("creating waypoint project")
	_, err = createWaypointProject(ctx, act, CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        waypointServerAddr,
		OrgID:                req.OrgId,
		ClusterInfo:          clusterInfo,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to create waypoint project: %w", err)
	}

	resp.ServerAddress = waypointServerAddr
	resp.SecretNamespace = w.cfg.WaypointBootstrapTokenNamespace
	resp.SecretName = getTokenSecretName(req.OrgId)
	resp.KubeClusterInfo = &serverv1.KubeClusterInfo{
		Id:             w.cfg.OrgsK8sClusterID,
		Endpoint:       w.cfg.OrgsK8sPublicEndpoint,
		CaData:         w.cfg.OrgsK8sCAData,
		TrustedRoleArn: w.cfg.OrgsK8sRoleArn,
	}
	l.Debug("finished signup", "response", resp)
	return resp, nil
}

// createWaypointProject executes an activity to create the waypoint project on the org's server
func createWaypointProject(
	ctx workflow.Context,
	act *Activities,
	req CreateWaypointProjectRequest,
) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint project activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateWaypointProject, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func createNamespace(ctx workflow.Context, act *Activities, cnr CreateNamespaceRequest) (CreateNamespaceResponse, error) {
	var resp CreateNamespaceResponse

	l := workflow.GetLogger(ctx)

	if err := cnr.validate(); err != nil {
		return resp, err
	}
	l.Debug("executing create namespace activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateNamespace, cnr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func installWaypointServer(
	ctx workflow.Context,
	act *Activities,
	iwr InstallWaypointServerRequest,
) (InstallWaypointServerResponse, error) {
	var resp InstallWaypointServerResponse

	l := workflow.GetLogger(ctx)

	if err := iwr.validate(); err != nil {
		return resp, err
	}
	l.Debug("executing install waypoint activity")
	fut := workflow.ExecuteActivity(ctx, act.InstallWaypointServer, iwr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func bootstrapWaypointServer(
	ctx workflow.Context,
	act *Activities,
	bwsr BootstrapWaypointServerRequest,
) (BootstrapWaypointServerResponse, error) {
	var resp BootstrapWaypointServerResponse

	l := workflow.GetLogger(ctx)

	if err := bwsr.validate(); err != nil {
		return resp, err
	}
	l.Debug("bootstrapping install waypoint server activity")
	fut := workflow.ExecuteActivity(ctx, act.BootstrapWaypointServer, bwsr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func pingWaypointServer(
	ctx workflow.Context,
	act *Activities,
	pwsr PingWaypointServerRequest,
) (PingWaypointServerResponse, error) {
	var resp PingWaypointServerResponse

	l := workflow.GetLogger(ctx)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		HeartbeatTimeout:    time.Second * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	l.Debug("executing ping waypoint server activity")
	fut := workflow.ExecuteActivity(ctx, act.PingWaypointServer, pwsr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
