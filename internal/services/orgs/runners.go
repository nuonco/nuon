package services

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *service) GetRunners(ctx context.Context, orgID string) (*orgsv1.GetRunnersResponse, error) {
	wpRunners, err := s.WaypointRepo.ListRunners(ctx)
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
