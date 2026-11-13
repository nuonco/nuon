package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appconfigupdated"
	installupdated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updated"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/patcher"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type PatchInstallConfigParams struct {
	ApprovalOption app.InstallApprovalOption `json:"approval_option"`
}

type UpdateInstallRequest struct {
	Name          string                    `json:"name"`
	Metadata      *helpers.InstallMetadata  `json:"metadata,omitempty"`
	InstallConfig *PatchInstallConfigParams `json:"install_config"`
}

func (c *UpdateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateInstall
// @Summary				update an install
// @Description.markdown	update_install.md
// @Param					install_id	path	string					true	"app ID"
// @Param					req			body	UpdateInstallRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Install
// @Router					/v1/installs/{install_id} [PATCH]
func (s *service) UpdateInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req UpdateInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.updateInstall(ctx, installID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &installupdated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &appconfigupdated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue reconcile-emitters signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) updateInstall(ctx context.Context, installID string, req *UpdateInstallRequest) (*app.Install, error) {
	// Scope every read and write to the caller's org so this handler stays safe even
	// if the installauth middleware is bypassed (NUO-1150).
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	currentInstall := app.Install{}
	res := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&currentInstall)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	// update or create install config
	if req.InstallConfig != nil {
		installConfig := app.InstallConfig{
			InstallID:      currentInstall.ID,
			OrgID:          orgID,
			ApprovalOption: req.InstallConfig.ApprovalOption,
		}
		res := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "deleted_at"},
				{Name: "install_id"},
			},
			// never rewrite a config row owned by another org
			Where: clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: clause.Column{Table: "install_configs", Name: "org_id"}, Value: orgID},
			}},
			DoUpdates: clause.AssignmentColumns([]string{
				"approval_option",
			}),
		}).
			Create(&installConfig)

		if res.Error != nil {
			return nil, fmt.Errorf("unable to upsert install config: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return nil, stderr.ErrNotFound{
				Err:         errors.New("install config not found"),
				Description: "install config not found",
			}
		}
	}

	updateObj := app.Install{Name: req.Name}
	if req.Metadata != nil {
		updateObj.Metadata = generics.ToHstore(map[string]string{
			"managed_by": req.Metadata.ManagedBy,
		})
	}

	res = s.db.WithContext(ctx).
		Scopes(scopes.WithPatcher(patcher.PatcherOptions{})).
		Model(&currentInstall).
		Where(app.Install{OrgID: orgID}).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("AppSandboxConfig").
		UpdateColumns(&updateObj)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &currentInstall, nil
}
