package nuon

import (
	"github.com/go-openapi/runtime"
	runtimeclient "github.com/go-openapi/runtime/client"
)

func (c *client) getOrgIDAuthInfo() runtime.ClientAuthInfoWriter {
	return runtimeclient.Compose(
		runtimeclient.APIKeyAuth("X-Nuon-Org-ID", "header", c.OrgID),
		c.getApiKeyAuthInfo(),
	)
}

func (c *client) getApiKeyAuthInfo() runtime.ClientAuthInfoWriter {
	return runtimeclient.BearerToken(c.APIToken)
}
