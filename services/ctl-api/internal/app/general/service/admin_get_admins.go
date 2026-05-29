package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						AdminGetAdmins
// @Summary				list all accounts with admin access.
// @Description.markdown	admin_get_admins.md
// @Param					offset	query	int	false	"offset of results to return"	Default(0)
// @Param					limit	query	int	false	"limit of results to return"	Default(10)
// @Param					page	query	int	false	"page number of results to return"	Default(0)
// @Tags					general/admin
// @Security				AdminEmail
// @Produce					json
// @Success				200	{array}	app.Account
// @Router					/v1/general/admin-get-admins [GET]
func (s *service) AdminGetAdmins(ctx *gin.Context) {
	admins, err := s.getAdmins(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get admins: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, admins)
}

func (s *service) getAdmins(ctx *gin.Context) ([]app.Account, error) {
	var admins []app.Account
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where(app.Account{IsAdmin: true}).
		Order("email ASC").
		Find(&admins)
	if res.Error != nil {
		return nil, res.Error
	}

	admins, err := db.HandlePaginatedResponse(ctx, admins)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return admins, nil
}
