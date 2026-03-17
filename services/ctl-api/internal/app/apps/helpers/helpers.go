package helpers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	Cfg         *internal.Config
	GhClient    *github.Client
	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	L           *zap.Logger
	VcsHelpers  *vcshelpers.Helpers
	QueueClient *queueclient.Client
}

type Helpers struct {
	cfg         *internal.Config
	ghClient    *github.Client
	db          *gorm.DB
	v           *validator.Validate
	l           *zap.Logger
	vcsHelpers  *vcshelpers.Helpers
	queueClient *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		v:           params.V,
		cfg:         params.Cfg,
		ghClient:    params.GhClient,
		db:          params.DB,
		l:           params.L,
		vcsHelpers:  params.VcsHelpers,
		queueClient: params.QueueClient,
	}
}

// VCSHelpers returns the VCS helpers for git source resolution and token creation.
func (h *Helpers) VCSHelpers() *vcshelpers.Helpers {
	return h.vcsHelpers
}

// EnsureComponentQueue creates a Temporal queue workflow for the given component if one does not
// already exist. It is idempotent — safe to call on every sync.
func (h *Helpers) EnsureComponentQueue(ctx context.Context, componentID string) error {
	var existing app.Queue
	if res := h.db.WithContext(ctx).First(&existing, "owner_id = ?", componentID); res.Error == nil {
		return nil
	}

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     componentID,
		OwnerType:   plugins.TableName(h.db, app.Component{}),
		Namespace:   "components",
		MaxInFlight: 1,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to create queue for component %s: %w", componentID, err)
	}
	return nil
}
