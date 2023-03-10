package deployments

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	deploymentsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/deployments/v1"
)

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[deploymentsv1.GetInfoRequest],
) (*connect.Response[deploymentsv1.GetInfoResponse], error) {
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

func (s *server) getInfo(ctx context.Context, req *deploymentsv1.GetInfoRequest, wkflows workflows.Repo) (*deploymentsv1.GetInfoResponse, error) {
	resp, err := wkflows.GetDeploymentProvisionResponse(ctx, req.OrgId, req.AppId, req.ComponentId, req.DeploymentId)
	if err != nil {
		return nil, fmt.Errorf("unable to get install provision response: %w", err)
	}

	prResp := resp.Response.GetDeploymentStart()
	if prResp == nil {
		return nil, fmt.Errorf("invalid response object")
	}

	return &deploymentsv1.GetInfoResponse{
		Id:       req.DeploymentId,
		Response: prResp,
	}, nil
}
