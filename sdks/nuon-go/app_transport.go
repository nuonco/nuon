package nuon

import (
	"compress/gzip"
	"fmt"
	"io"
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

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Handle gzip decompression if needed
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}

		// Replace the response body with decompressed reader
		resp.Body = &gzipReadCloser{
			reader: gzipReader,
			closer: resp.Body,
		}

		// Remove Content-Encoding header since we've decompressed
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
		resp.Uncompressed = true
	}

	return resp, nil
}

// gzipReadCloser wraps a gzip reader and ensures both the gzip reader and original body are closed
type gzipReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (g *gzipReadCloser) Read(p []byte) (n int, err error) {
	return g.reader.Read(p)
}

func (g *gzipReadCloser) Close() error {
	if rc, ok := g.reader.(io.Closer); ok {
		rc.Close()
	}
	return g.closer.Close()
}

func (c *client) SetOrgID(orgID string) {
	c.appTransport.orgID = orgID
}

func (c *client) SetClientVersion(version string) {
	c.appTransport.clientVersion = fmt.Sprintf("nuoncli:%s", version)
}
