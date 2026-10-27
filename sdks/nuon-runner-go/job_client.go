package nuonrunner

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"

	genclient "github.com/nuonco/nuon/sdks/nuon-runner-go/client"
)

type jobClientKey struct{}

// WithJobClient attaches a job-scoped client (same type, job token as auth) to ctx.
func WithJobClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, jobClientKey{}, client)
}

// ClientFromContext returns the job-scoped client if one was attached.
func ClientFromContext(ctx context.Context) (Client, bool) {
	client, ok := ctx.Value(jobClientKey{}).(Client)
	return client, ok && client != nil
}

// ForJob returns the same client shape authenticated with the job token.
func (c *client) ForJob(jobID, token string) (Client, error) {
	if jobID == "" || token == "" {
		return nil, fmt.Errorf("job client requires a job id and token")
	}
	apiURL, err := url.Parse(c.APIURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api url: %w", err)
	}
	next := *c
	next.APIToken = token
	next.APITokenFile = ""
	next.appTransport = &appTransport{
		authToken:     token,
		authTokenFile: "",
		orgID:         c.appTransport.orgID,
		clientVersion: c.appTransport.clientVersion,
		transport:     c.appTransport.transport,
	}
	next.httpClient = &http.Client{
		Transport: next.appTransport,
		Timeout:   c.RequestTimeout,
	}
	transport := httptransport.NewWithClient(apiURL.Host, apiURL.Path, []string{apiURL.Scheme}, next.httpClient)
	next.genClient = genclient.New(transport, nil)
	return &next, nil
}
