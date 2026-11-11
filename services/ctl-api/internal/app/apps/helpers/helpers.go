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
	"github.com/nuonco/nuon/pkg/config"
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

func (h *Helpers) WithDB(db *gorm.DB) *Helpers {
	helper := *h
	helper.db = db
	return &helper
}

// VCSHelpers returns the VCS helpers for git source resolution and token creation.
func (h *Helpers) VCSHelpers() *vcshelpers.Helpers {
	return h.vcsHelpers
}

// QueueClient returns the queue client for enqueueing app-owned signals.
func (h *Helpers) QueueClient() *queueclient.Client {
	return h.queueClient
}

// UploadCustomNestedStackTemplates uploads the config's custom nested stack
// templates and persists the resulting hashes and ready status. This runs inline
// with the config write so a stack generated from the config always finds its
// templates already in S3.
func (h *Helpers) UploadCustomNestedStackTemplates(ctx context.Context, stackConfig *app.AppStackConfig) error {
	if len(stackConfig.CustomNestedStacks) == 0 && !stackUINeedsUpload(stackConfig.UIDefinition) {
		return nil
	}

	// Deployments without a template bucket configured leave the stacks pending
	// rather than failing the config sync outright.
	if h.cfg.AWSCloudFormationStackTemplateBucket == "" {
		return nil
	}

	uploader, err := h.stackTemplateUploader()
	if err != nil {
		return err
	}

	if err := cloudformation.UploadCustomNestedStackTemplates(ctx, uploader, h.cfg.AWSCloudFormationStackTemplateBaseURL, stackConfig); err != nil {
		return err
	}

	columns := stackUploadColumns(stackConfig)
	if len(columns) == 0 {
		return nil
	}

	if res := h.db.WithContext(ctx).
		Model(stackConfig).
		Select(columns).
		Updates(stackConfig); res.Error != nil {
		return fmt.Errorf("unable to persist custom nested stack templates: %w", res.Error)
	}

	return nil
}

// UploadUIDefinition uploads one createUiDefinition and rewrites it in place.
// A missing bucket leaves the body pending instead of failing the config write.
func (h *Helpers) UploadUIDefinition(ctx context.Context, orgID, appID string, ui *config.UploadedUIDefinition) error {
	if ui != nil {
		if err := config.ValidateCreateUIDefinition(ui.Contents); err != nil {
			return err
		}
	}
	if !stackUINeedsUpload(ui) {
		return nil
	}
	if h.cfg.AWSCloudFormationStackTemplateBucket == "" {
		return nil
	}

	uploader, err := h.stackTemplateUploader()
	if err != nil {
		return err
	}
	return cloudformation.UploadUIDefinition(ctx, uploader, h.cfg.AWSCloudFormationStackTemplateBaseURL, orgID, appID, ui)
}

func (h *Helpers) stackTemplateUploader() (cloudformation.BlobUploader, error) {
	uploader, err := s3uploader.NewS3Uploader(h.v,
		s3uploader.WithBucketName(h.cfg.AWSCloudFormationStackTemplateBucket),
		s3uploader.WithCredentials(h.cfg.CFTemplateUploadCreds()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create s3 uploader: %w", err)
	}
	return uploader, nil
}

func stackUINeedsUpload(ui *config.UploadedUIDefinition) bool {
	return ui != nil && (ui.Contents != "" || ui.ContentsHash != "")
}

func stackUploadColumns(stackConfig *app.AppStackConfig) []string {
	var columns []string
	if len(stackConfig.CustomNestedStacks) > 0 {
		columns = append(columns, "custom_nested_stacks")
	}
	if stackConfig.UIDefinition != nil {
		columns = append(columns, "uidefinition")
	}
	return columns
}
