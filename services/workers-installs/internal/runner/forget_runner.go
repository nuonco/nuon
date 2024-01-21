package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/k8s"
)

type ForgetWaypointRunnerRequest struct {
	TokenSecretNamespace string           `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string           `json:"org_server_address" validate:"required"`
	OrgID                string           `json:"org_id" validate:"required"`
	InstallID            string           `json:"install_id" validate:"required"`
	ClusterInfo          kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (a ForgetWaypointRunnerRequest) validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

type ForgetWaypointRunnerResponse struct{}

func (a *Activities) ForgetWaypointRunner(ctx context.Context, req ForgetWaypointRunnerRequest) (ForgetWaypointRunnerResponse, error) {
	var resp ForgetWaypointRunnerResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	provider, err := k8s.New(a.v, k8s.WithConfig(k8s.Config{
		Address: req.OrgServerAddr,
		Token: k8s.Token{
			Namespace: req.TokenSecretNamespace,
			Name:      waypoint.DefaultTokenSecretName(req.OrgID),
			Key:       waypoint.DefaultTokenSecretKey,
		},
		ClusterInfo: &req.ClusterInfo,
	}))
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}

	client, err := provider.Fetch(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
	}

	if err := a.forgetWaypointRunner(ctx, client, req.InstallID); err != nil {
		return resp, fmt.Errorf("failed to forget waypoint runner: %w", err)
	}
	return resp, nil
}

func (a *Activities) forgetWaypointRunner(ctx context.Context, client gen.WaypointClient, installID string) error {
	req := &gen.ForgetRunnerRequest{
		RunnerId: installID,
	}

	_, err := client.ForgetRunner(ctx, req)
	return err
}
