package introspection

import (
	"github.com/gin-gonic/gin"
)

const nuonDescription = "Returns details about nuon built in values, by reading the environment."

func (s *svc) GetNuonHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("NUON")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: helmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: nuonDescription,
		Response:    resp,
	})
}
