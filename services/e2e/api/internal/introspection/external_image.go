package introspection

import (
	"github.com/gin-gonic/gin"
)

const ExternalImageDescription = "Returns details the external image component by reading the environment."

func (s *svc) GetExternalImageHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("EXTERNAL_IMAGE")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: ExternalImageDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: ExternalImageDescription,
		Response:    resp,
	})
}
