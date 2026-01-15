package nuonrunner

import (
	"fmt"
	"net/http"
)

// appTransport is a transport that injects our authentication token and org id into the api request
type appTransport struct {
	authToken     string
	orgID         string
	clientVersion string

	transport http.RoundTripper
}

func (t *appTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.authToken)
	if t.orgID != "" {
		req.Header.Set("X-Nuon-Org-ID", t.orgID)
	}
	if t.clientVersion != "" {
		req.Header.Set("X-Nuon-Client-Version", t.clientVersion)
	}

	return t.transport.RoundTrip(req)
}

func (c *client) SetOrgID(orgID string) {
	c.appTransport.orgID = orgID
}

func (c *client) SetClientVersion(version string) {
	c.appTransport.clientVersion = fmt.Sprintf("runner:%s", version)
}
