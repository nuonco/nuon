package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var defaultSupportUsers = []string{
	// Dre Smith
	"google-oauth2|113884954942864770921",
	//Jon Morehouse
	"google-oauth2|114670241124324496631",
	//Jordan Acosta
	"google-oauth2|106750268626571499868",
	//Nat Hamilton
	"google-oauth2|107796233904597398271",
	// Pavi Sandhu
	"google-oauth2|117375967099708763726",
	// Pete Lyons
	"google-oauth2|110347044904830192078",
}

func (s *service) CreateSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	for _, userID := range defaultSupportUsers {
		if err := s.createUser(ctx, orgID, userID); err != nil {
			ctx.Error(fmt.Errorf("unable to add users to org: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
