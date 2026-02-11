package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminRemoveTagsRequest struct {
	Tags []string `json:"tags" validate:"required,min=1,dive,required"`
}

func (r *AdminRemoveTagsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID                      AdminRemoveOrgTags
// @Summary                 remove tags from an organization
// @Description.markdown    admin_remove_org_tags.md
// @Param                   org_id  path    string  true    "org ID or name"
// @Tags                    orgs/admin
// @Security                AdminEmail
// @Accept                  json
// @Param                   req body    AdminRemoveTagsRequest  true    "Input"
// @Produce                 json
// @Success                 200 {object}    app.Org
// @Router                  /v1/orgs/{org_id}/admin-remove-tags [POST]
func (s *service) AdminRemoveTags(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminRemoveTagsRequest
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

	// Build set of tags to remove
	tagsToRemove := make(map[string]bool)
	for _, tag := range req.Tags {
		tagsToRemove[tag] = true
	}

	// Filter out tags to remove
	var newTags []string
	for _, tag := range org.Tags {
		if !tagsToRemove[tag] {
			newTags = append(newTags, tag)
		}
	}

	org.Tags = newTags

	// Update only the tags field
	if err := s.db.WithContext(ctx).Model(org).Select("tags").Updates(org).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update org tags: %w", err))
		return
	}

	// Reload org
	org, err = s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}
