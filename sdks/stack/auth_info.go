package stack

import (
	"github.com/go-openapi/runtime"
	runtimeclient "github.com/go-openapi/runtime/client"
)

func bearerAuth(token string) runtime.ClientAuthInfoWriter {
	return runtimeclient.Compose(runtimeclient.BearerToken(token))
}
