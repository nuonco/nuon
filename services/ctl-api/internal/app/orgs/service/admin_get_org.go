package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID						AdminGetOrg
// @Summary				get an org by name
// @Description.markdown	admin_get_org.md
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					name	query	string	false	"org name or id"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/admin-get [GET]
func (s *service) AdminGetOrg(ctx *gin.Context) {
	nameOrID := ctx.DefaultQuery("name", "")

	// Validate nameOrID is not empty
	if nameOrID == "" {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("name parameter is required"),
			Description: "name query parameter is required",
		})
		return
	}

	org, err := s.adminGetOrg(ctx, nameOrID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) adminGetOrg(ctx context.Context, nameOrID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		Unscoped().
		Preload("CreatedBy").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Where("name = ? OR name LIKE ? OR id = ?", nameOrID, "%"+nameOrID+"%", nameOrID).
		First(&org)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("org not found"),
				Description: fmt.Sprintf("org with name or id '%s' not found", nameOrID),
			}
		}
		return nil, fmt.Errorf("unable to get org by name or id: %w", res.Error)
	}

	return &org, nil
}
