package stacks

import (
	"context"
	"net/http"

	"github.com/nuonco/nuon/pkg/publicnet"
	"github.com/nuonco/nuon/pkg/publicnet/templatefetch"
	"github.com/nuonco/nuon/pkg/urlpolicy"
)

const MaxTemplateFetchBytes = templatefetch.MaxBytes

func NewTemplateFetchClient() *http.Client {
	return templatefetch.NewClient()
}

func NewLocalTemplateFetchClient() *http.Client {
	return templatefetch.NewLocalClient()
}

func GetTemplate(client *http.Client, policy *urlpolicy.Policy, rawURL string) (*http.Response, error) {
	ctx := publicnet.WithRedirectCheck(context.Background(), policy.RedirectCheck(urlpolicy.CategoryStackTemplates, "a stack template"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
