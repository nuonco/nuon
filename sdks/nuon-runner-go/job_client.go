package nuonrunner

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"

	genclient "github.com/nuonco/nuon/sdks/nuon-runner-go/client"
)

type jobScope struct {
	jobID string
	token string
}

type jobScopeKey struct{}
type jobClientKey struct{}

func WithJobScope(ctx context.Context, jobID, token string) context.Context {
	ctx = context.WithValue(ctx, jobScopeKey{}, jobScope{jobID: jobID, token: token})
	return ctx
}

func jobScopeFrom(ctx context.Context) (jobScope, bool) {
	scope, ok := ctx.Value(jobScopeKey{}).(jobScope)
	return scope, ok && scope.token != ""
}

func WithJobClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, jobClientKey{}, client)
}

func ClientFromContext(ctx context.Context) (Client, bool) {
	client, ok := ctx.Value(jobClientKey{}).(Client)
	return client, ok && client != nil
}

func (c *client) ForJob(jobID, token string) (Client, error) {
	if jobID == "" || token == "" {
		return nil, fmt.Errorf("job client requires a job id and token")
	}
	apiURL, err := url.Parse(c.APIURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api url: %w", err)
	}
	next := *c
	next.appTransport = &appTransport{
		authToken:     c.appTransport.authToken,
		authTokenFile: c.appTransport.authTokenFile,
		orgID:         c.appTransport.orgID,
		clientVersion: c.appTransport.clientVersion,
		transport:     c.appTransport.transport,
		jobID:         jobID,
		jobToken:      token,
	}
	next.httpClient = &http.Client{
		Transport: next.appTransport,
		Timeout:   c.RequestTimeout,
	}
	transport := httptransport.NewWithClient(apiURL.Host, apiURL.Path, []string{apiURL.Scheme}, next.httpClient)
	next.genClient = genclient.New(transport, nil)
	return &next, nil
}
