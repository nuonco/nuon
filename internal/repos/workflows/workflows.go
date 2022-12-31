package workflows

import (
	"context"

	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

const (
	requestFilename  string = "request.json"
	responseFilename string = "response.json"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=workflows_mock.go -source=workflows.go -package=workflows
type Repo interface {
	GetOrgProvisionRequest(ctx context.Context) (*sharedv1.Request, error)
	GetOrgProvisionResponse(ctx context.Context) (*sharedv1.Response, error)
}

// contextGetter is used to grab an org's context out of the passed in context
type contextGetter func(context.Context) (*orgcontext.Context, error)

type repo struct {
	ctxGetter contextGetter
}

var _ Repo = (*repo)(nil)

// New returns a default repo with the default orgcontext getter
func New() *repo {
	return &repo{
		ctxGetter: orgcontext.Get,
	}
}
