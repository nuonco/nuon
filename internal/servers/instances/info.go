package instances

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	instancesv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/instances/v1"
)

func ensureShortIDsGetInfoRequest(msg *instancesv1.GetInfoRequest) error {
	orgID, err := servers.EnsureShortID(msg.OrgId)
	if err != nil {
		return fmt.Errorf("invalid orgID: %w", err)
	}
	msg.OrgId = orgID

	appID, err := servers.EnsureShortID(msg.AppId)
	if err != nil {
		return fmt.Errorf("invalid appID: %w", err)
	}
	msg.AppId = appID

	componentID, err := servers.EnsureShortID(msg.ComponentId)
	if err != nil {
		return fmt.Errorf("invalid componentID: %w", err)
	}
	msg.ComponentId = componentID

	deploymentID, err := servers.EnsureShortID(msg.DeploymentId)
	if err != nil {
		return fmt.Errorf("invalid deploymentID: %w", err)
	}
	msg.DeploymentId = deploymentID

	installID, err := servers.EnsureShortID(msg.InstallId)
	if err != nil {
		return fmt.Errorf("invalid installID: %w", err)
	}
	msg.InstallId = installID

	return nil
}

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[instancesv1.GetInfoRequest],
) (*connect.Response[instancesv1.GetInfoResponse], error) {
	if err := ensureShortIDsGetInfoRequest(req.Msg); err != nil {
		return nil, fmt.Errorf("unable to ensure ids: %w", err)
	}

	wkflowsRepo, err := s.WorkflowsRepo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflows repo: %w", err)
	}

	resp, err := s.getInfo(ctx, req.Msg, wkflowsRepo)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return connect.NewResponse(resp), nil
}

func (s *server) getInfo(ctx context.Context, req *instancesv1.GetInfoRequest, wkflows workflows.Repo) (*instancesv1.GetInfoResponse, error) {
	resp, err := wkflows.GetInstanceProvisionResponse(ctx, req.OrgId, req.AppId, req.ComponentId, req.DeploymentId, req.InstallId)
	if err != nil {
		return nil, fmt.Errorf("unable to get install provision response: %w", err)
	}

	prResp := resp.Response.GetInstanceProvision()
	if prResp == nil {
		return nil, fmt.Errorf("invalid response object")
	}

	return &instancesv1.GetInfoResponse{
		Id:       req.InstallId,
		Response: prResp,
	}, nil
}
