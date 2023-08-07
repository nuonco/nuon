package introspection

import (
	"github.com/gin-gonic/gin"
)

const SecretsDescription = "Returns details about secrets, by reading the environment."

func (s *svc) GetSecretsHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("SECRET")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: HelmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: SecretsDescription,
		Response:    resp,
	})
}
