package introspection

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

const HelmDescription = "Returns details about the helm charts installed, and their values."

func (s *svc) GetHelmHandler(ctx *gin.Context) {
	resp, err := s.getHelmHandler(ctx)
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: HelmDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: HelmDescription,
		Response:    resp,
	})
}

type helmChartResponse struct {
	// Name is the name of the release
	Name string `json:"name,omitempty"`
	// Info provides information about a release
	Info *release.Info `json:"info,omitempty"`
	// Chart is the chart that was released.
	ChartMetadata *chart.Metadata `json:"chart_metadata,omitempty"`
	// Hooks are all of the hooks declared for this release.
	Hooks []*release.Hook `json:"hooks,omitempty"`
	// Version is an int which represents the revision of the release.
	Version int `json:"version,omitempty"`
	// Namespace is the kubernetes namespace of the release.
	Namespace string `json:"namespace,omitempty"`
	// Labels of the release.
	// Disabled encoding into Json cause labels are stored in storage driver metadata field.
	Labels map[string]string `json:"-"`
}

type helmResponse struct {
	Charts map[string]*helmChartResponse
}

func (s *svc) getHelmHandler(ctx context.Context) (*helmResponse, error) {
	resp := &helmResponse{
		Charts: make(map[string]*helmChartResponse, 0),
	}

	helmCfg, err := s.getHelmCfg(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get helm config: %w", err)
	}

	client := action.NewList(helmCfg)
	client.All = true
	client.AllNamespaces = true

	listResp, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("unable to get list response: %w", err)
	}
	for _, release := range listResp {
		k := fmt.Sprintf("%s.%s", release.Namespace, release.Name)
		resp.Charts[k] = &helmChartResponse{
			Name:          release.Name,
			Info:          release.Info,
			ChartMetadata: release.Chart.Metadata,
			Hooks:         release.Hooks,
			Version:       release.Version,
			Namespace:     release.Namespace,
			Labels:        release.Labels,
		}
	}

	return resp, nil
}
