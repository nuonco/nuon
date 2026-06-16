package activities

import (
	"context"

	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// GithubClient defines the GitHub API operations needed by VCS activities.
// This mirrors helpers.GithubClient to avoid an import cycle.
type GithubClient interface {
	GetInstallation(ctx context.Context, installID string) (*github.Installation, error)
	ListInstallationRepos(ctx context.Context, vcsConn *app.VCSConnection) ([]*github.Repository, error)
	CreateOrgWebhook(ctx context.Context, vcsConn *app.VCSConnection, webhookURL string, secret string) (int64, error)
}

type Params struct {
	fx.In

	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	GhClient    GithubClient
	QueueClient *queueclient.Client
	BlobService blobstore.Service
}

type Activities struct {
	cfg         *internal.Config
	db          *gorm.DB
	ghClient    GithubClient
	queueClient *queueclient.Client
	blobSvc     blobstore.Service
}

func New(params Params) *Activities {
	return &Activities{
		cfg:         params.Cfg,
		db:          params.DB,
		ghClient:    params.GhClient,
		queueClient: params.QueueClient,
		blobSvc:     params.BlobService,
	}
}
