package stack

import (
	"github.com/go-openapi/runtime"
	runtimeclient "github.com/go-openapi/runtime/client"
)

// bearerAuth adapts a token resolved by sdks/auth to the writer the generated
// client expects. Token precedence stays in sdks/auth.
func bearerAuth(token string) runtime.ClientAuthInfoWriter {
	return runtimeclient.Compose(runtimeclient.BearerToken(token))
}
