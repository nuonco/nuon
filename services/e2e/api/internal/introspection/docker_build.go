package introspection

import (
	"github.com/gin-gonic/gin"
)

const DockerBuildDescription = "Returns details the docker build component by reading the environment."

func (s *svc) GetDockerBuildHandler(ctx *gin.Context) {
	resp, err := s.getEnvByPrefix("DOCKER_BUILD")
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: DockerBuildDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: DockerBuildDescription,
		Response:    resp,
	})
}
