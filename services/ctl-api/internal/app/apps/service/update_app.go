package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateAppRequest struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	DisplayName     string            `json:"display_name"`
	SlackWebhookURL *string           `json:"slack_webhook_url"`
	ConfigRepo      *string           `json:"config_repo"`
	ConfigDirectory *string           `json:"config_directory"`
	LabelColors     map[string]string `json:"label_colors,omitempty"`
}

func (c *UpdateAppRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateApp
// @Summary				update an app
// @Description.markdown	update_app.md
// @Param					app_id	path	string				true	"app ID"
// @Param					req		body	UpdateAppRequest	true	"Input"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.App
// @Router					/v1/apps/{app_id} [patch]
func (s *service) UpdateApp(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}

	var req UpdateAppRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.Response{}, stderr.NewInvalidRequest(err)
	}
	if err := req.Validate(s.v); err != nil {
		return require.Response{}, fmt.Errorf("invalid request: %w", err)
	}

	target := access.Target()
	tx := access.DB()
	var currentApp app.App
	if err := tx.Where(app.App{ID: target.ID, OrgID: target.OrgID}).Take(&currentApp).Error; err != nil {
		return require.Response{}, err
	}
	if err := s.updateApp(tx, target, &currentApp, &req); err != nil {
		return require.Response{}, err
	}
	reader := *s
	reader.db = tx
	result, err := reader.loadAppDetails(tx.Statement.Context, target.OrgID, tx.Where(app.App{ID: target.ID}))
	return require.Response{Status: http.StatusOK, Body: result}, err
}

func (s *service) updateApp(db *gorm.DB, target permissions.ResolvedTarget, currentApp *app.App, req *UpdateAppRequest) error {
	updates := app.App{
		Name:        req.Name,
		Description: generics.NewNullString(req.Description),
		DisplayName: generics.NewNullString(req.DisplayName),
	}

	if req.ConfigRepo != nil && req.ConfigDirectory != nil {
		updates.ConfigRepo = *req.ConfigRepo
		updates.ConfigDirectory = *req.ConfigDirectory
	}

	res := db.
		Model(&app.App{}).
		Where(app.App{ID: target.AppID, OrgID: target.OrgID}).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update app: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("app not found %s %w", target.AppID, gorm.ErrRecordNotFound)
	}
	currentApp.Name = updates.Name
	if req.Description != "" {
		currentApp.Description = updates.Description
	}
	if req.DisplayName != "" {
		currentApp.DisplayName = updates.DisplayName
	}

	if req.SlackWebhookURL != nil {
		res = db.
			Select("slack_webhook_url").
			Model(&app.NotificationsConfig{}).
			Where(&app.NotificationsConfig{
				OwnerID: target.AppID,
			}).
			Updates(app.NotificationsConfig{
				SlackWebhookURL: *req.SlackWebhookURL,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to sync app notifications config: %w", res.Error)
		}
	}

	if req.LabelColors != nil {
		res = db.
			Select("label_colors").
			Model(&app.App{}).
			Where(app.App{ID: target.AppID, OrgID: target.OrgID}).
			Updates(app.App{
				LabelColors: labels.Labels(req.LabelColors),
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update app label colors: %w", res.Error)
		}
	}

	return nil
}
