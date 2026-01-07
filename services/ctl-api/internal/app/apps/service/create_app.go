package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppRequest struct {
	Name            string `json:"name" validate:"required,entity_name"`
	Description     string `json:"description"`
	DisplayName     string `json:"display_name"`
	SlackWebhookURL string `json:"slack_webhook_url"`
}

func (c *CreateAppRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateApp
// @Summary				create an app
// @Description.markdown	create_app.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.App
// @Router					/v1/apps [post]
func (s *service) CreateApp(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	user, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createApp(ctx, user, org, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	// Update user journey for first app creation
	if err := s.accountsHelpers.UpdateUserJourneyStepForFirstAppCreate(ctx, user.ID, app.ID); err != nil {
		// Log error but don't fail app creation
		s.l.Warn("failed to update user journey for first app creation",
			zap.String("account_id", user.ID),
			zap.String("app_id", app.ID),
			zap.Error(err))
	}

	s.evClient.Send(ctx, app.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})
	s.evClient.Send(ctx, app.ID, &signals.Signal{
		Type: signals.OperationPollDependencies,
	})
	s.evClient.Send(ctx, app.ID, &signals.Signal{
		Type: signals.OperationProvision,
	})
	ctx.JSON(http.StatusCreated, app)
}

func (s *service) createApp(ctx context.Context, acct *app.Account, org *app.Org, req *CreateAppRequest) (*app.App, error) {
	// Check if queues feature is enabled
	queuesEnabled := org.Features[string(app.OrgFeatureQueues)]

	// Start transaction if queues are enabled
	var tx *gorm.DB
	if queuesEnabled {
		tx = s.db.WithContext(ctx).Begin()
	} else {
		tx = s.db.WithContext(ctx)
	}

	newApp := app.App{
		OrgID:             org.ID,
		Name:              req.Name,
		Description:       generics.NewNullString(req.Description),
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision app",
		DisplayName:       generics.NewNullString(req.DisplayName),
	}
	newApp.NotificationsConfig = app.NotificationsConfig{
		EnableSlackNotifications: slices.Contains([]app.AccountType{app.AccountTypeAuth0, app.AccountTypeAuth}, acct.AccountType),
		EnableEmailNotifications: slices.Contains([]app.AccountType{app.AccountTypeAuth0, app.AccountTypeAuth}, acct.AccountType),
		InternalSlackWebhookURL:  org.NotificationsConfig.InternalSlackWebhookURL,
		SlackWebhookURL:          req.SlackWebhookURL,
	}

	if err := tx.Create(&newApp).Error; err != nil {
		if queuesEnabled {
			tx.Rollback()
		}
		return nil, fmt.Errorf("unable to create app: %w", err)
	}

	// Create queue if feature is enabled
	if queuesEnabled {
		queue, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     newApp.ID,
			OwnerType:   "App",
			Namespace:   "apps",
			MaxInFlight: 3,
			MaxDepth:    50,
		})
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("unable to create queue for app: %w", err)
		}

		// Update app with queue ID
		newApp.QueueID = queue.ID
		if err := tx.Save(&newApp).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("unable to update app with queue_id: %w", err)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("unable to commit transaction: %w", err)
		}
	}

	return &newApp, nil
}
