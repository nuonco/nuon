package users

import (
	"context"

	"github.com/bufbuild/connect-go"
	userv1 "github.com/powertoolsdev/protos/api/generated/types/user/v1"
)

func (s *server) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[userv1.GetCurrentUserRequest],
) (*connect.Response[userv1.GetCurrentUserResponse], error) {
	res := connect.NewResponse(&userv1.GetCurrentUserResponse{})
	return res, nil
}

func (s *server) GetUser(
	ctx context.Context,
	req *connect.Request[userv1.GetUserRequest],
) (*connect.Response[userv1.GetUserResponse], error) {
	res := connect.NewResponse(&userv1.GetUserResponse{})
	return res, nil
}

func (s *server) GetUsersByOrg(
	ctx context.Context,
	req *connect.Request[userv1.GetUsersByOrgRequest],
) (*connect.Response[userv1.GetUsersByOrgResponse], error) {
	res := connect.NewResponse(&userv1.GetUsersByOrgResponse{})
	return res, nil
}
