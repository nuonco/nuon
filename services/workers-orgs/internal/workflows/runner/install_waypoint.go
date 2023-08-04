package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypointhelm "github.com/powertoolsdev/mono/pkg/waypoint/helm"
	"go.temporal.io/sdk/activity"
	"helm.sh/helm/v3/pkg/release"
)

func runnerServiceAccountName(orgID string) string {
	return fmt.Sprintf("waypoint-static-runner-%s", orgID)
}

func runnerOdrServiceAccountName(orgID string) string {
	return fmt.Sprintf("waypoint-odr-%s", orgID)
}

type InstallWaypointRequest struct {
	Namespace   string      `json:"namespace" validate:"required"`
	ReleaseName string      `json:"release_name" validate:"required"`
	Chart       *helm.Chart `json:"chart" validate:"required"`
	Atomic      bool        `json:"atomic"`

	RunnerConfig RunnerConfig     `json:"runner_config" validate:"required"`
	OrgID        string           `json:"org_id" validate:"required"`
	ClusterInfo  kube.ClusterInfo `json:"cluster_info" validate:"required"`

	// These are exposed for testing. Do not use otherwise
	CreateNamespace bool `json:"create_namespace"`
}

func (i InstallWaypointRequest) validate() error {
	validate := validator.New()
	return validate.Struct(i)
}

type InstallWaypointResponse struct{}

type installer interface {
	Install(context.Context, *helm.InstallConfig) (*release.Release, error)
}

// getWaypointRunnerValues returns the set of values needed to configure the request
func getWaypointRunnerValues(req InstallWaypointRequest) (map[string]interface{}, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	values := waypointhelm.NewDefaultOrgRunnerValues()
	values.Runner.ID = req.RunnerConfig.ID
	values.Runner.Server.Addr = req.RunnerConfig.ServerAddr
	values.Runner.Server.TLS = true
	values.Runner.Server.TLSSkipVerify = true
	values.Runner.Server.Cookie = req.RunnerConfig.Cookie
	values.Runner.Odr.ServiceAccount.Create = true
	values.Runner.Odr.ServiceAccount.Name = runnerOdrServiceAccountName(req.OrgID)
	values.Runner.Odr.ServiceAccount.Annotations = map[string]string{
		"eks.amazonaws.com/role-arn": req.RunnerConfig.OdrIAMRoleArn,
	}

	values.Runner.ServiceAccount.Create = true
	values.Runner.ServiceAccount.Name = runnerServiceAccountName(req.OrgID)
	values.Bootstrap.ServiceAccount.Create = false

	var vals map[string]interface{}
	err := mapstructure.Decode(values, &vals)
	return vals, err
}

// TODO(jdt): make this idempotent
func (a *Activities) InstallWaypoint(
	ctx context.Context,
	req InstallWaypointRequest,
) (InstallWaypointResponse, error) {
	resp := InstallWaypointResponse{}

	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	l := activity.GetLogger(ctx)

	kCfg, err := a.getKubeConfig(&req.ClusterInfo)
	if err != nil {
		return resp, fmt.Errorf("unable to get kube config: %w", err)
	}

	vals, err := getWaypointRunnerValues(req)
	if err != nil {
		return resp, fmt.Errorf("failed to create helm values: %w", err)
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
		return resp, fmt.Errorf("failed to install: %w", err)
	}

	l.Debug("finished installing waypoint", "response", resp)
	return resp, nil
}
