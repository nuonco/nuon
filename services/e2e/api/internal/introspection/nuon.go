package introspection

import (
	"github.com/gin-gonic/gin"
)

const NuonDescription = "Returns details about nuon built in values, by reading the environment."

func (s *svc) GetNuonHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("NUON")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: HelmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: NuonDescription,
		Response:    resp,
	})
}
