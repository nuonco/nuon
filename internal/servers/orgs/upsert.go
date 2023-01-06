package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers"
	"github.com/powertoolsdev/go-generics"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) UpsertOrg(
	ctx context.Context,
	req *connect.Request[orgv1.UpsertOrgRequest],
) (*connect.Response[orgv1.UpsertOrgResponse], error) {
	org, err := s.Svc.UpsertOrg(ctx, models.OrgInput{
		ID:      generics.ToPtr(req.Msg.Id),
		Name:    req.Msg.Name,
		OwnerID: req.Msg.OwnerId,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to upsert org: %w", err)
	}

	res := connect.NewResponse(&orgv1.UpsertOrgResponse{
		Org: &orgv1.Org{
			Id:        org.ID.String(),
			Name:      org.Name,
			OwnerId:   org.CreatedByID,
			UpdatedAt: servers.TimeToDatetime(org.UpdatedAt),
			CreatedAt: servers.TimeToDatetime(org.CreatedAt),
		},
	})
	return res, nil
}
