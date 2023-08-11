package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AllOrgsResponse []GetOrgResponse

func (s *service) GetAllOrgs(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, []GetOrgResponse{
		{
			ID: "inlabc",
		},
	})
}
