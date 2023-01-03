package orgs

import (
	"context"

	"github.com/powertoolsdev/api/internal/repos"
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

func (s *server) GetOrgsByUser(context.Context, *orgv1.GetOrgsByUserRequest) (*orgv1.GetOrgsByUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrgsByUser not implemented")
}
func (s *server) UpsertOrg(context.Context, *orgv1.UpsertOrgRequest) (*orgv1.UpsertOrgResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertOrg not implemented")
}
