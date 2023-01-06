package users

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	userv1 "github.com/powertoolsdev/protos/api/generated/types/user/v1"
)

func (s *server) UpsertOrgMember(
	ctx context.Context,
	req *connect.Request[userv1.UpsertOrgMemberRequest],
) (*connect.Response[userv1.UpsertOrgMemberResponse], error) {
	userOrg, err := s.Svc.UpsertUserOrg(ctx, models.UserOrgInput{
		UserID: req.Msg.UserId,
		OrgID:  req.Msg.OrgId,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&userv1.UpsertOrgMemberResponse{
		OrgMember: &userv1.OrgMember{
			UserId: userOrg.UserID,
			OrgId:  userOrg.OrgID.String(),
		},
	}), nil
}
