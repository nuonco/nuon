package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	runnerhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V                *validator.Validate
	Helpers          *helpers.Helpers
	DB               *gorm.DB `name:"psql"`
	BlobService      blobstore.Service
	AcctClient       *account.Client
	AuthzClient      *authz.Client
	Cfg              *internal.Config
	L                *zap.Logger
	RunnerHelpers    *runnerhelpers.Helpers
	VCSHelpers       *vcshelpers.Helpers
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	RunbooksHelpers  *runbookshelpers.Helpers
	InstallHelpers   *installhelpers.Helpers
	QueueClient      *queueclient.Client
}

type Activities struct {
	v                *validator.Validate
	db               *gorm.DB
	helpers          *helpers.Helpers
	blobSvc          blobstore.Service
	acctClient       *account.Client
	authzClient      *authz.Client
	cfg              *internal.Config
	l                *zap.Logger
	runnerHelpers    *runnerhelpers.Helpers
	vcsHelpers       *vcshelpers.Helpers
	componentHelpers *componenthelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	runbooksHelpers  *runbookshelpers.Helpers
	installHelpers   *installhelpers.Helpers
	queueClient      *queueclient.Client
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:                params.V,
		db:               params.DB,
		helpers:          params.Helpers,
		blobSvc:          params.BlobService,
		acctClient:       params.AcctClient,
		authzClient:      params.AuthzClient,
		cfg:              params.Cfg,
		l:                params.L,
		runnerHelpers:    params.RunnerHelpers,
		vcsHelpers:       params.VCSHelpers,
		componentHelpers: params.ComponentHelpers,
		actionsHelpers:   params.ActionsHelpers,
		runbooksHelpers:  params.RunbooksHelpers,
		installHelpers:   params.InstallHelpers,
		queueClient:      params.QueueClient,
	}, nil
}
