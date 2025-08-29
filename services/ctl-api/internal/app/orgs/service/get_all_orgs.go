package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID			AdminGetAllOrgs
// @BasePath	/v1/orgs
// @Summary	Return all orgs
// @Schemes
// @Description	return all orgs
// @Param			type						query	string	false	"type of orgs to return"		Default(real)
// @Param			offset						query	int		false	"offset of results to return"	Default(0)
// @Param			limit						query	int		false	"limit of results to return"	Default(10)
// @Param			page						query	int		false	"page number of results to return"	Default(0)
// @Tags			orgs/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array}	app.Org
// @Router			/v1/orgs [GET]
func (s *service) GetAllOrgs(ctx *gin.Context) {
	orgTyp := ctx.DefaultQuery("type", "real")

	orgs, err := s.getAllOrgs(ctx, orgTyp)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getAllOrgs(ctx *gin.Context, typ string) ([]*app.Org, error) {
	var orgs []*app.Org

	where := app.Org{}
	if !generics.StringOneOf(typ, "", "all") {
		where.OrgType = app.OrgType(typ)
	}

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Joins("JOIN accounts ON accounts.id=orgs.created_by_id").
		Where(where).
		Order("orgs.created_at desc").
		Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all orgs: %w", res.Error)
	}

	orgs, err := db.HandlePaginatedResponse(ctx, orgs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return orgs, nil
}
