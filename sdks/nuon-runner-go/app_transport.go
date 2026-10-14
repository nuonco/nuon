package nuonrunner

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// appTransport is a transport that injects our authentication token and org id into the api request
type appTransport struct {
	authToken     string
	authTokenFile string
	orgID         string
	clientVersion string

	transport http.RoundTripper
}

func (t *appTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token := t.authToken
	if t.authTokenFile != "" {
		file, err := os.Open(t.authTokenFile)
		if err != nil {
			return nil, fmt.Errorf("open auth token file: %w", err)
		}
		contents, err := io.ReadAll(io.LimitReader(file, 16*1024+1))
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("read auth token file: %w", err)
		}
		token = strings.TrimSpace(string(contents))
		if token == "" || len(contents) > 16*1024 || strings.ContainsAny(token, " \t\r\n\x00") {
			return nil, fmt.Errorf("auth token file contains an invalid credential")
		}
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+token)
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
