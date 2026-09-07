package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

type Service struct {
	cfg         *config.Config
	allowWrites bool
	endpoint    string
	name        string
}

type Option func(*Service)

func WithEndpoint(endpoint string) Option {
	return func(s *Service) {
		s.endpoint = endpoint
	}
}

func WithName(name string) Option {
	return func(s *Service) {
		s.name = name
	}
}

func New(cfg *config.Config, allowWrites bool, opts ...Option) *Service {
	svc := &Service{cfg: cfg, allowWrites: allowWrites}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func (s *Service) Run(ctx context.Context) error {
	upstream, err := s.connectUpstream(ctx)
	if err != nil {
		return fmt.Errorf("connecting to upstream MCP: %w", err)
	}
	defer upstream.Close()

	server, err := s.buildProxyServer(ctx, upstream)
	if err != nil {
		return fmt.Errorf("building proxy server: %w", err)
	}

	return server.Run(ctx, &mcp.StdioTransport{})
}

func (s *Service) mcpEndpoint() (string, error) {
	if s.endpoint != "" {
		return s.endpoint, nil
	}
	return EndpointFromAPIURL(s.cfg.APIURL)
}

func EndpointFromAPIURL(apiURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(apiURL, "/"))
	if err != nil || parsed.Hostname() == "" {
		return "", fmt.Errorf("unable to derive MCP URL from API URL %q; pass --url", apiURL)
	}

	if IsLocalAPIURL(apiURL) {
		return "http://localhost:8088/mcp", nil
	}

	if host := parsed.Hostname(); !strings.HasPrefix(host, "api.") {
		return "", fmt.Errorf("unable to derive MCP URL from API URL %q: hostname must start with api.; pass --url", apiURL)
	}

	parsed.Host = strings.Replace(parsed.Host, "api.", "mcp.", 1)
	parsed.Path = "/mcp"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func IsLocalAPIURL(apiURL string) bool {
	parsed, err := url.Parse(strings.TrimRight(apiURL, "/"))
	if err != nil {
		return false
	}

	switch parsed.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func NameFromAPIURL(apiURL string) string {
	switch strings.TrimRight(apiURL, "/") {
	case "https://api.nuon.co":
		return "nuon"
	case "https://api.stage.nuon.co":
		return "nuon-stage"
	default:
		return "nuon-local"
	}
}

func (s *Service) serverName() string {
	if s.name != "" {
		return s.name
	}
	return NameFromAPIURL(s.cfg.APIURL)
}

func (s *Service) connectUpstream(ctx context.Context) (*mcp.ClientSession, error) {
	endpoint, err := s.mcpEndpoint()
	if err != nil {
		return nil, err
	}

	transport := &mcp.StreamableClientTransport{
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Transport: &authRoundTripper{
				token: s.cfg.APIToken,
				orgID: s.cfg.OrgID,
				base:  http.DefaultTransport,
			},
		},
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "nuon-cli-proxy",
		Version: version.Version,
	}, nil)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) buildProxyServer(ctx context.Context, upstream *mcp.ClientSession) (*mcp.Server, error) {
	res, err := upstream.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing upstream tools: %w", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    s.serverName(),
		Version: version.Version,
	}, nil)

	for _, tool := range res.Tools {
		if !s.allowWrites && isWriteTool(tool) {
			continue
		}

		toolCopy := *tool
		mcp.AddTool(server, &toolCopy, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			result, err := upstream.CallTool(ctx, &mcp.CallToolParams{
				Name:      req.Params.Name,
				Arguments: req.Params.Arguments,
			})
			if err != nil {
				return nil, nil, err
			}
			return result, nil, nil
		})
	}

	return server, nil
}

func isWriteTool(tool *mcp.Tool) bool {
	return strings.HasPrefix(tool.Description, "WRITE OPERATION:")
}

type authRoundTripper struct {
	token string
	orgID string
	base  http.RoundTripper
}

func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
	if a.orgID != "" {
		req.Header.Set("X-Nuon-Org-ID", a.orgID)
	}
	return a.base.RoundTrip(req)
}
