package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	appsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/apps/v1"
)

func ensureShortIDsGetInfoRequest(msg *appsv1.GetInfoRequest) error {
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

	return nil
}

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[appsv1.GetInfoRequest],
) (*connect.Response[appsv1.GetInfoResponse], error) {
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
