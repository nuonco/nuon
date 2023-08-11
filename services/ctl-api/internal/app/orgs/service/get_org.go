package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetOrgResponse struct {
	ID string `json:"id"`
}

func (s *service) GetOrg(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &GetOrgResponse{
		ID: "inlabc",
	})
}
