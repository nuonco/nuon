package introspection

import (
	"github.com/gin-gonic/gin"
)

const SandboxDescription = "Returns details about the sandbox, by reading the environment."

func (s *svc) GetSandboxHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("SANDBOX")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: SandboxDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: SandboxDescription,
		Response:    resp,
	})
}
