package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminAddPriorityRequest struct {
	HighPriority    bool `json:"high_priority" default:"false"`
	DefaultPriority bool `json:"low_priority" default:"false"`

	InternalOnly bool `json:"internal_only" default:"false"`
}

func (a AdminAddPriorityRequest) getPriority() int {
	if a.HighPriority {
		return 9999
	}

	if a.DefaultPriority {
		return 10
	}

	if a.InternalOnly {
		return -1
	}

	return 0
}

//	@ID						AdminAddPriority
//	@Summary				set priority on an org
//	@Description.markdown	add_org_priority.md
//	@Param					org_id	path	string	true	"org ID for your current org"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	AdminAddPriorityRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/{org_id}/admin-add-priority [POST]
func (s *service) AdminAddPriority(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminAddPriorityRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.setOrgPriority(ctx, orgID, req.getPriority()); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) setOrgPriority(ctx context.Context, orgID string, priority int) error {
	org := app.Org{
		ID: orgID,
	}

	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		Priority: priority,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to set org priority: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
