package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type WaitlistResponse struct {
	ID             string    `json:"id,omitzero"`
	OrgName        string    `json:"org_name,omitzero"`
	CreatedByID    string    `json:"created_by_id,omitzero"`
	CreatedByEmail string    `json:"created_by_email,omitzero"`
	CreatedAt      time.Time `json:"created_at,omitzero"`
}

// @ID						AdminGetWaitlist
// @Summary				get waitlist
// @Description.markdown	admin_get_waitlist.md
// @Tags					general/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{array}	WaitlistResponse
// @Router					/v1/general/waitlist [GET]
func (s *service) AdminGetWaitlist(ctx *gin.Context) {
	waitlist, err := s.adminGetWaitlist(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, waitlist)
}

func (s *service) adminGetWaitlist(ctx context.Context) ([]WaitlistResponse, error) {
	waitlist := []*app.Waitlist{}
	waitlistResponse := []WaitlistResponse{}
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Order("created_at desc").
		Find(&waitlist)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	for _, w := range waitlist {
		waitlistResponse = append(waitlistResponse, WaitlistResponse{
			ID:             w.ID,
			OrgName:        w.OrgName,
			CreatedByID:    w.CreatedByID,
			CreatedByEmail: w.CreatedBy.Email,
			CreatedAt:      w.CreatedAt,
		})
	}

	return waitlistResponse, nil
}
