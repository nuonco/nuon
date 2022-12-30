package workflows

import (
	"context"

	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=workflows_mock.go -source=workflows.go -package=workflows
type Repo interface {
	GetOrgProvisionRequest(ctx context.Context) (*sharedv1.Request, error)
	GetOrgProvisionResponse(ctx context.Context) (*sharedv1.Response, error)
}

type repo struct{}

var _ Repo = (*repo)(nil)
