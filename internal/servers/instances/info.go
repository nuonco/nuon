package instances

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	instancesv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/instances/v1"
)

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[instancesv1.GetInfoRequest],
) (*connect.Response[instancesv1.GetInfoResponse], error) {
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
