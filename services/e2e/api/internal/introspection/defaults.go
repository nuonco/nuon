package introspection

import (
	"github.com/gin-gonic/gin"
)

const DefaultsDescription = "Returns details about default values, by reading the environment."

func (s *svc) GetDefaultsHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("DEFAULT")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: HelmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: DefaultsDescription,
		Response:    resp,
	})
}
