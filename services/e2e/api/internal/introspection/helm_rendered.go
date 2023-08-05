package introspection

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

const helmRenderedDescription = "Returns the rendered manifests for a helm deploy."

func (s *svc) GetHelmRenderedHandler(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	resp, err := s.getHelmRendered(ctx, namespace, name)
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: helmRenderedDescription,
			Err:         err,
		})
		return
	}

	ctx.String(http.StatusOK, resp)
}

func (s *svc) getHelmRendered(ctx context.Context, namespace, name string) (string, error) {
	helmCfg, err := s.getHelmCfg(ctx, namespace)
	if err != nil {
		return "", fmt.Errorf("unable to get helm config: %w", err)
	}

	client := action.NewList(helmCfg)
	client.All = true
	client.AllNamespaces = true

	listResp, err := client.Run()
	if err != nil {
		return "", fmt.Errorf("unable to get list response: %w", err)
	}

	var release *release.Release
	for _, release = range listResp {
		if release.Name == name {
			break
		}
	}
	if release == nil {
		return "", fmt.Errorf("release %s not found", name)
	}

	return release.Manifest, nil
}
