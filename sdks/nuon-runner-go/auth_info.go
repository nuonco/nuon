package nuonrunner

import (
	"github.com/go-openapi/runtime"
	runtimeclient "github.com/go-openapi/runtime/client"
)

func (c *client) getAuthInfo() runtime.ClientAuthInfoWriter {
	return runtimeclient.Compose(
		c.getApiKeyAuthInfo(),
	)
}

func (c *client) getApiKeyAuthInfo() runtime.ClientAuthInfoWriter {
	if c.appTransport != nil {
		return runtimeclient.BearerToken(c.appTransport.currentAuthToken())
	}
	return runtimeclient.BearerToken(c.APIToken)
}
