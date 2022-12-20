package orgs

import (
	"context"

	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/request"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type server struct {
	repo repos.OrgRepo
}

func NewOrgServer(db *gorm.DB) server {
	return server{
		repo: repos.NewOrgRepo(db),
	}
}

var _ orgv1.OrgsServiceServer

func (s *server) GetOrg(ctx context.Context, req *orgv1.GetOrgRequest) (*orgv1.GetOrgResponse, error) {
	orgID, err := request.ParseID(req.OrgId)
	if err != nil {
		return nil, err
	}

	org, err := s.repo.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}

	orgProto, err := orgModelToProto(org)
	if err != nil {
		return nil, err
	}

	return &orgv1.GetOrgResponse{
		Org: orgProto,
	}, nil
}
func (s *server) GetOrgsByUser(context.Context, *orgv1.GetOrgsByUserRequest) (*orgv1.GetOrgsByUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrgsByUser not implemented")
}
func (s *server) UpsertOrg(context.Context, *orgv1.UpsertOrgRequest) (*orgv1.UpsertOrgResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertOrg not implemented")
}
func (s *server) DeleteOrg(context.Context, *orgv1.DeleteOrgRequest) (*orgv1.DeleteOrgResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteOrg not implemented")
}
