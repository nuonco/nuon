package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

var defaultSupportUsers = [][2]string{
	// Dre Smith
	{"google-oauth2|113884954942864770921", "andre@nuon.co"},
	// Jon Morehouse
	{"google-oauth2|114670241124324496631", "jon@nuon.co"},
	// Jordan Acosta
	{"google-oauth2|106750268626571499868", "jordan@nuon.co"},
	// Nat Hamilton
	{"google-oauth2|107796233904597398271", "nat@nuon.co"},
	// Sam Boyer
	{"google-oauth2|112612105639694013121", "sam@nuon.co"},
	// Rob Bruce
	{"google-oauth2|111073782439789496112", "rob@nuon.co"},
	// Fred Diego
	{"google-oauth2|101745198892864595751", "fred@nuon.co"},
	// Casey Bierman
	{"google-oauth2|102295167972372616803", "casey@nuon.co"},
}

// @ID AdminCreateSupportUsers
// @BasePath	/v1/orgs
// @Summary	Add nuon users as support members
// @Description.markdown create_org_support_users.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-support-users [POST]
func (s *service) CreateSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	middlewares.SetAccountIDGinContext(ctx, org.CreatedByID)
	for _, user := range defaultSupportUsers {
		if err := s.createSupportUser(ctx, user[0], user[1], orgID); err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (s *service) createSupportUser(ctx context.Context, email, subject, orgID string) error {
	acct, err := s.authzClient.FindAccount(ctx, email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		acct, err = s.authzClient.CreateAccount(ctx, email, subject)
		if err != nil {
			return err
		}
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, orgID, acct.ID); err != nil {
		return err
	}

	return nil
}
