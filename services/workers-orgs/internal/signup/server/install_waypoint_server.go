package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/helm/waypoint"
	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	"helm.sh/helm/v3/pkg/release"
)

type InstallWaypointServerRequest struct {
	Namespace   string      `json:"namespace" validate:"required"`
	ReleaseName string      `json:"release_name" validate:"required"`
	Chart       *helm.Chart `json:"chart" validate:"required"`
	Atomic      bool        `json:"atomic"`

	ClusterInfo kube.ClusterInfo `json:"cluster_info" validate:"required"`

	// These are exposed for testing. Do not use otherwise
	CreateNamespace bool `json:"create_namespace"`
}

func (i InstallWaypointServerRequest) validate() error {
	validate := validator.New()
	return validate.Struct(i)
}

type InstallWaypointServerResponse struct{}

type installer interface {
	Install(context.Context, *helm.InstallConfig) (*release.Release, error)
}

// TODO(jdt): make this idempotent
func (a *Activities) InstallWaypointServer(ctx context.Context, req InstallWaypointServerRequest) (InstallWaypointServerResponse, error) {
	resp := InstallWaypointServerResponse{}

	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	l := activity.GetLogger(ctx)

	values := waypoint.NewDefaultOrgServerValues()
	var vals map[string]interface{}
	if err := mapstructure.Decode(values, &vals); err != nil {
		return resp, fmt.Errorf("failed to convert helm values: %w", err)
	}

	var err error
	kCfg := a.Kubeconfig
	if kCfg == nil {
		kCfg, err = kube.ConfigForCluster(&req.ClusterInfo)
		if err != nil {
			return resp, fmt.Errorf("failed to get config for cluster: %w", err)
		}
	}

	cfg := &helm.InstallConfig{
		Namespace:       req.Namespace,
		ReleaseName:     req.ReleaseName,
		Chart:           req.Chart,
		Atomic:          req.Atomic,
		Values:          vals,
		CreateNamespace: req.CreateNamespace,
		Kubeconfig:      kCfg,
		Logger:          l,
	}
	_, err = a.helmInstaller.Install(ctx, cfg)
	if err != nil {
		return resp, fmt.Errorf("failed to install waypoint: %w", err)
	}

	l.Debug("finished installing waypoint", "response", resp)
	return resp, nil
}

var (
	ErrInvalidReleaseName = errors.New("invalid release name")
	ErrInvalidChart       = errors.New("invalid chart")
)
