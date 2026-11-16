package cloudformation

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Templates struct {
	cfg *internal.Config

	// fetchClient overrides templateFetchClient; see WithTemplateFetchClient.
	fetchClient *http.Client
}

type Params struct {
	fx.In

	Cfg *internal.Config
}

func NewTemplates(params Params) *Templates {
	return &Templates{
		cfg: params.Cfg,
	}
}
