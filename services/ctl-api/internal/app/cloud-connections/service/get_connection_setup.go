package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID GetCloudConnectionSetup
// @Summary get cloud connection setup material
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Param repository query []string false "Transient exact ECR ARN, GAR resource, or ACR repository path for read-only grant generation; not stored" collectionFormat(multi)
// @Param existing_principal query bool false "Reuse an existing principal in Terraform/CloudFormation; AWS trust contribution must be merged in the original owning role. CLI always detects and reuses existing identities." default(false)
// @Success 200 {object} SetupResponse
// @Router /v1/cloud-connections/{connection_id}/setup [get]
func (s *service) Setup(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	repositories := ctx.QueryArray("repository")
	for _, repository := range repositories {
		if err := cloudconnections.ValidateImageRepository(connection, repository); err != nil {
			ctx.Error(userError(err))
			return
		}
	}
	var query struct {
		ExistingPrincipal bool `form:"existing_principal"`
	}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.Error(userError(err))
		return
	}
	ctx.JSON(http.StatusOK, s.setupWithPrincipal(connection, query.ExistingPrincipal, repositories...))
}
