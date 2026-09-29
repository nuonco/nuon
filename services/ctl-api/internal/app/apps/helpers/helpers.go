package helpers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/aws/s3uploader"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
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

func (h *Helpers) VCSHelpers() *vcshelpers.Helpers {
	return h.vcsHelpers
}

func (h *Helpers) QueueClient() *queueclient.Client {
	return h.queueClient
}

func (h *Helpers) UploadCustomNestedStackTemplates(ctx context.Context, db *gorm.DB, stackConfig *app.AppStackConfig) error {
	if len(stackConfig.CustomNestedStacks) == 0 {
		return nil
	}

	if h.cfg.AWSCloudFormationStackTemplateBucket == "" {
		return nil
	}

	uploader, err := s3uploader.NewS3Uploader(h.v,
		s3uploader.WithBucketName(h.cfg.AWSCloudFormationStackTemplateBucket),
		s3uploader.WithCredentials(h.cfg.CFTemplateUploadCreds()),
	)
	if err != nil {
		return fmt.Errorf("unable to create s3 uploader: %w", err)
	}

	if err := cloudformation.UploadCustomNestedStackTemplates(ctx, uploader, h.cfg.AWSCloudFormationStackTemplateBaseURL, stackConfig); err != nil {
		return err
	}

	if res := db.WithContext(ctx).
		Model(stackConfig).
		Select("custom_nested_stacks").
		Updates(stackConfig); res.Error != nil {
		return fmt.Errorf("unable to persist custom nested stack templates: %w", res.Error)
	}

	return nil
}
