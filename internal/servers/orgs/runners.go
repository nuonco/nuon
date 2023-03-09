package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func ensureShortIDsGetRunnersRequest(msg *orgsv1.GetRunnersRequest) error {
	orgID, err := servers.EnsureShortID(msg.OrgId)
	if err != nil {
		return fmt.Errorf("invalid orgID: %w", err)
	}
	msg.OrgId = orgID

	return nil
}

func (s *server) GetRunners(
	ctx context.Context,
	req *connect.Request[orgsv1.GetRunnersRequest],
) (*connect.Response[orgsv1.GetRunnersResponse], error) {
	if err := ensureShortIDsGetRunnersRequest(req.Msg); err != nil {
		return nil, fmt.Errorf("unable to ensure ids: %w", err)
	}

	wpRepo, err := s.WaypointRepo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get waypoint repo: %w", err)
	}

	resp, err := s.getRunners(ctx, wpRepo)
	if err != nil {
		return nil, fmt.Errorf("unable to get runners: %w", err)
	}

	return connect.NewResponse(resp), nil
}

func (s *server) getRunners(
	ctx context.Context,
	wpRepo waypoint.Repo,
) (*orgsv1.GetRunnersResponse, error) {
	wpRunners, err := wpRepo.ListRunners(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list runners: %w", err)
	}

	runners := make([]*orgsv1.RunnerInfo, len(wpRunners.Runners))
	for idx, wpRunner := range wpRunners.Runners {
		runners[idx] = &orgsv1.RunnerInfo{
			Id:            wpRunner.Id,
			Kind:          fmt.Sprintf("%s", wpRunner.Kind),
			Labels:        wpRunner.Labels,
			Online:        wpRunner.Online,
			AdoptionState: wpRunner.AdoptionState.String(),
			FirstSeen:     waypoint.TimestampToDatetime(wpRunner.FirstSeen),
			LastSeen:      waypoint.TimestampToDatetime(wpRunner.LastSeen),
		}
	}

	return &orgsv1.GetRunnersResponse{
		Runners: runners,
	}, nil
}
