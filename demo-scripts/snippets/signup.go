package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
)

type SignupRequest struct {
	Name       string `json:"name"`
	IAMRoleARN string `json:"iam_role_arn"`
	Region     string `json:"region"`
}

func (s *service) CreateApp(ctx *gin.Context) {
	var req SignupRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	nuonAPI, err := nuon.New(s.v,
		nuon.WithAuthToken("your-api-token"),
		nuon.WithOrgID("your-org-id"),
	)
	if err != nil {
		ctx.Error(fmt.Errorf("internal api error"))
		return
	}

	install, err := nuonAPI.CreateInstall(ctx, "your-app-id", &models.ServiceCreateInstallRequest{
		AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
			IamRoleArn: &req.IAMRoleARN,
			Region:     req.Region,
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("internal api error"))
		return
	}

	ctx.JSON(http.StatusCreated, install)
}
