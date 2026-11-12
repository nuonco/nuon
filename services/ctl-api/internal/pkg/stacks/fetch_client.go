package stacks

import (
	"net/http"

	"github.com/nuonco/nuon/pkg/publicnet/templatefetch"
)

const MaxTemplateFetchBytes = templatefetch.MaxBytes

func NewTemplateFetchClient() *http.Client {
	return templatefetch.NewClient()
}

func NewLocalTemplateFetchClient() *http.Client {
	return templatefetch.NewLocalClient()
}
