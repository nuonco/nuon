package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateOrgRequest struct {
	Name string `json:"name" validate:"required"`
}

func (c *UpdateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateOrg
// @Summary				Update current org
// @Description.markdown	update_org.md
// @Param					req	body	UpdateOrgRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/current [PATCH]
func (s *service) UpdateOrg(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}

	var req UpdateOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}
	if err := req.Validate(s.v); err != nil {
		return require.Response{}, fmt.Errorf("invalid request: %w", err)
	}

	org, err := s.updateOrg(access.DB(), access.Target().OrgID, &req)
	if err != nil {
		return require.Response{}, err
	}

	return require.Response{Status: http.StatusOK, Body: org}, nil
}

func (s *service) updateOrg(db *gorm.DB, orgID string, req *UpdateOrgRequest) (*app.Org, error) {
	filter := app.Org{ID: orgID}
	res := db.Model(&app.Org{}).Where(filter).Update("name", req.Name)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("org not found: %w", gorm.ErrRecordNotFound)
	}
	var org app.Org
	if err := db.Where(filter).First(&org).Error; err != nil {
		return nil, fmt.Errorf("unable to reload org: %w", err)
	}

	return &org, nil
}
