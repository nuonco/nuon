package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"go.temporal.io/sdk/activity"
	"helm.sh/helm/v3/pkg/release"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypointhelm "github.com/powertoolsdev/mono/pkg/waypoint/helm"
)

type RunnerConfig struct {
	ID            string `validate:"required" json:"id"`
	Cookie        string `validate:"required" json:"cookie"`
	ServerAddr    string `validate:"required" json:"server_addr"`
	OdrIAMRoleArn string `json:"odr_iam_role_arn"`
}

type InstallWaypointRequest struct {
	InstallID    string              `validate:"required" json:"install_id"`
	Namespace    string              `validate:"required" json:"namespace"`
	ReleaseName  string              `validate:"required" json:"release_name"`
	Chart        *helm.Chart         `validate:"required" json:"chart"`
	Atomic       bool                `json:"atomic"`
	ClusterInfo  kube.ClusterInfo    `validate:"required" json:"cluster_info"`
	RunnerConfig RunnerConfig        `validate:"required" json:"runner_config"`
	Auth         *credentials.Config `json:"auth"`

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

	values := waypointhelm.NewDefaultInstallValues()
	values.Runner.Enabled = true
	values.Runner.ID = req.RunnerConfig.ID
	values.Runner.Server.Addr = req.RunnerConfig.ServerAddr
	values.Runner.Server.TLS = true
	values.Runner.Server.TLSSkipVerify = true
	values.Runner.Server.Cookie = req.RunnerConfig.Cookie

	values.Server.Enabled = false
	values.UI.Service.Enabled = false
	values.Bootstrap.ServiceAccount.Create = false

	values.Runner.Odr.ServiceAccount.Create = true
	values.Runner.Odr.ServiceAccount.Name = runnerOdrServiceAccountName(req.InstallID)
	values.Runner.Odr.ServiceAccount.Annotations = map[string]string{
		"eks.amazonaws.com/role-arn": req.RunnerConfig.OdrIAMRoleArn,
	}

	values.Runner.ServiceAccount.Create = true
	values.Runner.ServiceAccount.Name = runnerServiceAccountName(req.InstallID)

	var vals map[string]interface{}
	err := mapstructure.Decode(values, &vals)
	return vals, err
}

// TODO(jdt): make this idempotent
func (a *Activities) InstallWaypoint(ctx context.Context, req InstallWaypointRequest) (InstallWaypointResponse, error) {
	resp := InstallWaypointResponse{}

	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	l := activity.GetLogger(ctx)

	var err error
	if (req.Auth != nil) && (*req.Auth != credentials.Config{}) {
		envVars, err := credentials.FetchEnv(ctx, req.Auth)
		if err != nil {
			return resp, fmt.Errorf("unable to get credentials: %w", err)
		}
		req.ClusterInfo.EnvVars = envVars

	}

	kCfg, err := kube.ConfigForCluster(&req.ClusterInfo)
	if err != nil {
		return resp, fmt.Errorf("failed to get config for cluster: %w", err)
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
