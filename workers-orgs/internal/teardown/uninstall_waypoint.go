package teardown

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/go-helm"
	"go.temporal.io/sdk/activity"
	"helm.sh/helm/v3/pkg/release"
)

type UninstallWaypointRequest struct {
	Namespace   string
	ReleaseName string
}

type UninstallWaypointResponse struct{}

type uninstaller interface {
	Uninstall(context.Context, *helm.UninstallConfig) (*release.UninstallReleaseResponse, error)
}

// TODO(jdt): make this idempotent
func (a *Activities) UninstallWaypoint(ctx context.Context, req UninstallWaypointRequest) (UninstallWaypointResponse, error) {
	resp := UninstallWaypointResponse{}
	l := activity.GetLogger(ctx)

	if err := validateUninstallWaypointRequest(req); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	cfg := &helm.UninstallConfig{
		Namespace:   req.Namespace,
		ReleaseName: req.ReleaseName,
		Kubeconfig:  a.Kubeconfig,
		Logger:      l,
	}

	_, err := a.helmUninstaller.Uninstall(ctx, cfg)
	if err != nil {
		return resp, fmt.Errorf("failed to uninstall: %w", err)
	}

	l.Debug("finished uninstalling waypoint", "response", resp)
	return resp, nil
}

var (
	ErrInvalidReleaseName = errors.New("invalid release name")
)

func validateUninstallWaypointRequest(req UninstallWaypointRequest) error {
	if req.Namespace == "" {
		return fmt.Errorf("%w: namespace must be specified", ErrInvalidNamespaceName)
	}
	if req.ReleaseName == "" {
		return fmt.Errorf("%w: release name must be specified", ErrInvalidReleaseName)
	}

	return nil
}
