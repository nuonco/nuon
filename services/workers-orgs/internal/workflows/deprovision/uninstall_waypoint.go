package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	"helm.sh/helm/v3/pkg/release"
)

type UninstallWaypointRequest struct {
	Namespace   string           `validate:"required"`
	ReleaseName string           `validate:"required"`
	ClusterInfo kube.ClusterInfo `validate:"required"`
}

func (r UninstallWaypointRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type UninstallWaypointResponse struct{}

type uninstaller interface {
	Uninstall(context.Context, *helm.UninstallConfig) (*release.UninstallReleaseResponse, error)
}

func (a *Activities) UninstallWaypoint(ctx context.Context, req UninstallWaypointRequest) (UninstallWaypointResponse, error) {
	resp := UninstallWaypointResponse{}
	l := activity.GetLogger(ctx)
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	kCfg, err := a.getKubeConfig(&req.ClusterInfo)
	if err != nil {
		return resp, fmt.Errorf("unable to get kube config: %w", err)
	}

	cfg := &helm.UninstallConfig{
		Namespace:   req.Namespace,
		ReleaseName: req.ReleaseName,
		Kubeconfig:  kCfg,
		Logger:      l,
	}
	_, err = a.helmUninstaller.Uninstall(ctx, cfg)
	if err != nil {
		return resp, fmt.Errorf("failed to uninstall: %w", err)
	}

	l.Debug("finished uninstalling waypoint", "response", resp)
	return resp, nil
}
