package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	appsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/apps/v1"
)

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[appsv1.GetInfoRequest],
) (*connect.Response[appsv1.GetInfoResponse], error) {
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

func (s *server) getInfo(ctx context.Context, req *appsv1.GetInfoRequest, wkflows workflows.Repo) (*appsv1.GetInfoResponse, error) {
	resp, err := wkflows.GetAppProvisionResponse(ctx, req.OrgId, req.AppId)
	if err != nil {
		return nil, fmt.Errorf("unable to get app provision response: %w", err)
	}

	prResp := resp.Response.GetAppsProvision()
	if prResp == nil {
		return nil, fmt.Errorf("invalid response object")
	}

	return &appsv1.GetInfoResponse{
		Id:         req.AppId,
		Repository: prResp.Repository,
	}, nil
}
