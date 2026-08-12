package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateRoleRequest struct {
	// Title is the role's display name.
	Title string `json:"title"`
	// Description explains what the role grants.
	Description *string `json:"description"`
	// Contexts names the assignment surfaces the role may be offered on; an
	// empty list withdraws the role from every picker.
	Contexts *[]string `json:"contexts"`
	// Permissions replace the role's scoped permission entries when provided.
	Permissions []PermissionEntryRequest `json:"permissions"`
}

// @ID						UpdateRole
// @Summary				Update a custom role for the current org
// @Description			Update a custom role's metadata or replace its scoped permission entries. Managed roles cannot be edited.
// @Param					role_id	path	string				true	"role ID"
// @Param					req		body	UpdateRoleRequest	true	"Input"
// @Tags					roles
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Role
// @Router					/v1/roles/{role_id} [PATCH]
func (s *service) UpdateRole(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(userErr(fmt.Errorf("unable to parse request: %w", err)))
		return
	}

	role, err := s.requireCustomRole(ctx, org.ID, ctx.Param("role_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	// Contexts is a jsonb column written through a gorm field serializer, which
	// only runs for struct updates — a map value reaches the driver as a raw Go
	// slice and Postgres rejects it as a record. Select carries the fields the
	// request actually set, so a zero value still writes.
	var updates app.Role
	var fields []string
	if strings.TrimSpace(req.Title) != "" {
		title := strings.TrimSpace(req.Title)
		if err := s.requireAvailableTitle(ctx, org.ID, title, role.ID); err != nil {
			ctx.Error(err)
			return
		}
		updates.Title = title
		fields = append(fields, "title")
	}
	if req.Description != nil {
		updates.Description = *req.Description
		fields = append(fields, "description")
	}
	if req.Contexts != nil {
		if err := validateContexts(*req.Contexts); err != nil {
			ctx.Error(userErr(err))
			return
		}
		updates.Contexts = *req.Contexts
		fields = append(fields, "contexts")
	}

	var entries []app.PermissionEntry
	if req.Permissions != nil {
		entries, err = s.validatePermissionEntries(ctx, org, req.Permissions)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(fields) > 0 {
			if err := tx.Model(&app.Role{}).Where(app.Role{ID: role.ID}).Select(fields).Updates(updates).Error; err != nil {
				return fmt.Errorf("unable to update role: %w", err)
			}
		}

		if entries != nil {
			if err := tx.Model(&app.Policy{}).
				Where(app.Policy{RoleID: role.ID}).
				Updates(app.Policy{ScopedPermissions: entries}).Error; err != nil {
				return fmt.Errorf("unable to update role permissions: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	role, err = s.getOrgRole(ctx, org.ID, role.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, role)
}
