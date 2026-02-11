package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminAddTagsRequest struct {
	Tags []string `json:"tags" validate:"required,min=1,dive,required"`
}

func (r *AdminAddTagsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID                      AdminAddOrgTags
// @Summary                 add tags to an organization
// @Description.markdown    admin_add_org_tags.md
// @Param                   org_id  path    string  true    "org ID or name"
// @Tags                    orgs/admin
// @Security                AdminEmail
// @Accept                  json
// @Param                   req body    AdminAddTagsRequest true    "Input"
// @Produce                 json
// @Success                 200 {object}    app.Org
// @Router                  /v1/orgs/{org_id}/admin-add-tags [POST]
func (s *service) AdminAddTags(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminAddTagsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Build set of existing tags for uniqueness check
	existingTags := make(map[string]bool)
	for _, tag := range org.Tags {
		existingTags[tag] = true
	}

	// Add only unique tags
	for _, tag := range req.Tags {
		if !existingTags[tag] {
			org.Tags = append(org.Tags, tag)
		}
	}

	// Update only the tags field
	if err := s.db.WithContext(ctx).Model(org).Select("tags").Updates(org).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update org tags: %w", err))
		return
	}

	// Reload org to get fresh data with all preloads
	org, err = s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}
