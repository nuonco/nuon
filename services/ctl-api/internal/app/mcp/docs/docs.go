package docs

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

const (
	upstreamSearchTool = "search_nuon_documentation"
	upstreamFSTool     = "query_docs_filesystem_nuon_documentation"

	callTimeout = 30 * time.Second
)

func Instructions(endpoint string) string {
	return fmt.Sprintf("For how Nuon works (concepts, app config reference, CLI, guides, troubleshooting), call search_docs and read_docs_page before answering from memory or reading Nuon source code. Use this server's other tools for the state of the user's own org. To browse the docs directly, the user can also add the Nuon docs MCP at %s.", endpoint)
}

var pagePathRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9/_-]*$`)

type proxy struct {
	endpoint   string
	client     *mcp.Client
	httpClient *http.Client
}

func Register(server *mcp.Server, endpoint string) {
	p := &proxy{
		endpoint:   endpoint,
		client:     mcp.NewClient(&mcp.Implementation{Name: "nuon-ctl-docs", Version: "1.0.0"}, nil),
		httpClient: &http.Client{Timeout: callTimeout},
	}

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"search_docs",
		"Search Nuon docs",
		"Search the Nuon documentation (docs.nuon.co) for concepts, app config reference, CLI usage, guides, and troubleshooting. Returns matching sections with title, link, and page path. Pass the page path to read_docs_page for the full page.",
	), p.searchDocs)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"read_docs_page",
		"Read Nuon docs page",
		"Read the full markdown of one Nuon docs page by its page path (the Page value from search_docs, e.g. guides/agents/setup) or its docs.nuon.co URL.",
	), p.readDocsPage)
}

type searchDocsInput struct {
	Query string `json:"query" jsonschema:"what to look for, in plain language or keywords"`
}

func (p *proxy) searchDocs(ctx context.Context, _ *mcp.CallToolRequest, in searchDocsInput) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.Query) == "" {
		return nil, nil, fmt.Errorf("query is required")
	}
	return p.call(ctx, upstreamSearchTool, map[string]any{"query": in.Query})
}

type readDocsPageInput struct {
	Path string `json:"path" jsonschema:"page path from search_docs (e.g. guides/agents/setup) or a docs.nuon.co URL"`
}

func (p *proxy) readDocsPage(ctx context.Context, _ *mcp.CallToolRequest, in readDocsPageInput) (*mcp.CallToolResult, any, error) {
	path, err := normalizePagePath(in.Path)
	if err != nil {
		return nil, nil, err
	}
	return p.call(ctx, upstreamFSTool, map[string]any{"command": "cat /" + path + ".mdx"})
}

func normalizePagePath(raw string) (string, error) {
	path := strings.TrimSpace(raw)
	path = strings.TrimPrefix(path, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "docs.nuon.co")
	if i := strings.IndexAny(path, "#?"); i >= 0 {
		path = path[:i]
	}
	path = strings.TrimSuffix(path, ".mdx")
	path = strings.Trim(path, "/")

	if !pagePathRe.MatchString(path) {
		return "", fmt.Errorf("invalid docs page path %q: pass the Page value from search_docs, e.g. guides/agents/setup", raw)
	}
	return path, nil
}

func (p *proxy) call(ctx context.Context, tool string, args map[string]any) (*mcp.CallToolResult, any, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	session, err := p.client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             p.endpoint,
		HTTPClient:           p.httpClient,
		MaxRetries:           -1,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to reach Nuon docs at %s: %w", p.endpoint, err)
	}
	defer session.Close()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query Nuon docs: %w", err)
	}

	return &mcp.CallToolResult{Content: res.Content, IsError: res.IsError}, nil, nil
}
