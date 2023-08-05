package introspection

import (
	"github.com/gin-gonic/gin"
)

const terraformDescription = "Returns details about a connected terraform component, by reading the environment."

func (s *svc) GetTerraformHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("TERRAFORM_")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: helmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: terraformDescription,
		Response:    resp,
	})
}
