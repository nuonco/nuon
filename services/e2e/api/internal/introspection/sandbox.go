package introspection

import (
	"github.com/gin-gonic/gin"
)

const sandboxDescription = "Returns details about the sandbox, by reading the environment."

func (s *svc) GetSandboxHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("SANDBOX")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: helmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: sandboxDescription,
		Response:    resp,
	})
}
