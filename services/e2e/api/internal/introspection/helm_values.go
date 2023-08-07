package introspection

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/action"
)

const HelmValuesDescription = "Returns the final values for a helm chart deployment"

func (s *svc) GetHelmValuesHandler(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	resp, err := s.getHelmValues(ctx, namespace, name)
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: HelmValuesDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: HelmValuesDescription,
		Response:    resp,
	})
}

func (s *svc) getHelmValues(ctx context.Context, namespace, name string) (map[string]interface{}, error) {
	helmCfg, err := s.getHelmCfg(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm config: %w", err)
	}

	client := action.NewGetValues(helmCfg)
	client.AllValues = true

	resp, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm values: %w", err)
	}

	return resp, nil
}
