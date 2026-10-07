package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/pkg/agentclient"
)

type Service struct {
	cfg         *config.Config
	allowWrites bool
	endpoint    string
	name        string
	progress    progressRelay
}

// progressRelay forwards upstream progress notifications to the stdio client
// that started the tool call. Keyed by the progress token on that call.
type progressRelay struct {
	mu       sync.Mutex
	handlers map[string]func(context.Context, *mcp.ProgressNotificationParams)
}

func (r *progressRelay) listen(token any, fn func(context.Context, *mcp.ProgressNotificationParams)) func() {
	if token == nil {
		return func() {}
	}
	key := fmt.Sprint(token)
	r.mu.Lock()
	if r.handlers == nil {
		r.handlers = map[string]func(context.Context, *mcp.ProgressNotificationParams){}
	}
	r.handlers[key] = fn
	r.mu.Unlock()
	return func() {
		r.mu.Lock()
		delete(r.handlers, key)
		r.mu.Unlock()
	}
}

func (r *progressRelay) forward(ctx context.Context, req *mcp.ProgressNotificationClientRequest) {
	if req == nil || req.Params == nil || req.Params.ProgressToken == nil {
		return
	}
	r.mu.Lock()
	fn := r.handlers[fmt.Sprint(req.Params.ProgressToken)]
	r.mu.Unlock()
	if fn != nil {
		fn(ctx, req.Params)
	}
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

	// api.<host> is the public API. app.<host> is the dashboard, which proxies
	// /v1, so login accepts it as api_url (for example https://app.nuon.co).
	// Replacing that first label maps both onto mcp.<host>.
	prefix, ok := mcpHostPrefix(parsed.Hostname())
	if !ok {
		return "", fmt.Errorf("unable to derive MCP URL from API URL %q: hostname must start with api. or app.; pass --url", apiURL)
	}

	parsed.Host = strings.Replace(parsed.Host, prefix, "mcp.", 1)
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

func mcpHostPrefix(host string) (string, bool) {
	switch {
	case strings.HasPrefix(host, "api."):
		return "api.", true
	case strings.HasPrefix(host, "app."):
		return "app.", true
	default:
		return "", false
	}
}

func NameFromAPIURL(apiURL string) string {
	switch strings.TrimRight(apiURL, "/") {
	case "https://api.nuon.co", "https://app.nuon.co":
		return "nuon"
	case "https://api.stage.nuon.co", "https://app.stage.nuon.co":
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

	agent := s.cfg.Agent
	userAgent := ""
	command := ""
	if agent != "" {
		userAgent = agentclient.UserAgent(version.Version, agent)
		command = "nuon agents mcp"
	}
	transport := &mcp.StreamableClientTransport{
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Transport: &authRoundTripper{
				token:     s.cfg.APIToken,
				orgID:     s.cfg.OrgID,
				agent:     agent,
				userAgent: userAgent,
				command:   command,
				base:      http.DefaultTransport,
			},
		},
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "nuon-cli-proxy",
		Version: version.Version,
	}, &mcp.ClientOptions{
		ProgressNotificationHandler: s.progress.forward,
	})

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

	// Carry upstream's Instructions (e.g. the skills pointer) onto the local
	// proxy server — it's only sent to clients on initialize, so it doesn't
	// show up in ListTools/ListResources and must be forwarded explicitly.
	var instructions string
	if init := upstream.InitializeResult(); init != nil {
		instructions = init.Instructions
	}
	instructions = appendCLIInstructions(instructions, CLIBinary(), CLIConfigFlag(s.cfg))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    s.serverName(),
		Version: version.Version,
	}, &mcp.ServerOptions{
		Instructions: instructions,
	})

	for _, tool := range res.Tools {
		if !s.allowWrites && isWriteTool(tool) {
			continue
		}

		toolCopy := *tool
		mcp.AddTool(server, &toolCopy, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			params := &mcp.CallToolParams{
				Name:      req.Params.Name,
				Arguments: req.Params.Arguments,
			}
			if token := req.Params.GetProgressToken(); token != nil {
				params.SetProgressToken(token)
				defer s.progress.listen(token, func(ctx context.Context, p *mcp.ProgressNotificationParams) {
					_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
						ProgressToken: token,
						Message:       p.Message,
						Progress:      p.Progress,
						Total:         p.Total,
					})
				})()
			}
			result, err := upstream.CallTool(ctx, params)
			if err != nil {
				return nil, nil, err
			}
			return result, nil, nil
		})
	}

	if err := s.proxyResources(ctx, upstream, server); err != nil {
		return nil, fmt.Errorf("listing upstream resources: %w", err)
	}

	return server, nil
}

// proxyResources mirrors upstream's resources and resource templates (e.g.
// agent skills) onto the local server, forwarding reads back upstream.
// Resources are read-only content, so they're never subject to the
// --allow-writes filter applied to tools.
func (s *Service) proxyResources(ctx context.Context, upstream *mcp.ClientSession, server *mcp.Server) error {
	resources, err := upstream.ListResources(ctx, nil)
	if err != nil {
		return err
	}
	for _, resource := range resources.Resources {
		resourceCopy := *resource
		server.AddResource(&resourceCopy, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return upstream.ReadResource(ctx, &mcp.ReadResourceParams{URI: req.Params.URI})
		})
	}

	templates, err := upstream.ListResourceTemplates(ctx, nil)
	if err != nil {
		return err
	}
	for _, tmpl := range templates.ResourceTemplates {
		tmplCopy := *tmpl
		server.AddResourceTemplate(&tmplCopy, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return upstream.ReadResource(ctx, &mcp.ReadResourceParams{URI: req.Params.URI})
		})
	}

	return nil
}

func isWriteTool(tool *mcp.Tool) bool {
	return strings.HasPrefix(tool.Description, "WRITE OPERATION:")
}

// CLIBinary is the command name used to start this process, such as nuon or nuon-dev.
func CLIBinary() string {
	return cliBinary(os.Args)
}

func cliBinary(args []string) string {
	if len(args) == 0 {
		return "nuon"
	}
	base := filepath.Base(args[0])
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "nuon"
	}
	return base
}

// CLIConfigFlag is the -C value this process was started with. Empty when
// the MCP server or CLI command was not given -C.
func CLIConfigFlag(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	return cfg.ConfigFlag
}

// CLICommandInstructions tells an agent which Nuon binary to use for local CLI
// work this MCP server asks for. It includes -C only when that flag was passed.
func CLICommandInstructions(binary, configFlag string) string {
	name := commandName(binary)
	if configFlag == "" {
		return fmt.Sprintf(
			"When you run a local Nuon CLI command on behalf of this MCP server, use the %s binary that started it. Do not apply that binary to Nuon CLI commands for other work. Validate an app config directory with `%s apps validate` from that directory, or pass the directory as the argument. Do not upload config files to validate them.",
			name, name,
		)
	}
	quoted := strconv.Quote(configFlag)
	return fmt.Sprintf(
		"When you run a local Nuon CLI command on behalf of this MCP server, use %s and pass -C %s, matching how this server was started. Do not apply that binary or -C to Nuon CLI commands for other work. Validate an app config directory with `%s -C %s apps validate` from that directory, or pass the directory as the argument. Do not upload config files to validate them.",
		name, quoted, name, quoted,
	)
}

func commandName(binary string) string {
	if binary == "" {
		return "nuon"
	}
	if strings.ContainsAny(binary, " \t\"'\\") {
		return strconv.Quote(binary)
	}
	return binary
}

func appendCLIInstructions(upstream, binary, configFlag string) string {
	cli := CLICommandInstructions(binary, configFlag)
	if strings.TrimSpace(upstream) == "" {
		return cli
	}
	return upstream + " " + cli
}

type authRoundTripper struct {
	token     string
	orgID     string
	agent     string
	userAgent string
	command   string
	base      http.RoundTripper
}

func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
	if a.orgID != "" {
		req.Header.Set("X-Nuon-Org-ID", a.orgID)
	}
	if a.agent != "" {
		req.Header.Set(agentclient.Header, a.agent)
	}
	if a.userAgent != "" {
		req.Header.Set("User-Agent", a.userAgent)
	}
	if a.command != "" {
		req.Header.Set(agentclient.CommandHeader, a.command)
	}
	return a.base.RoundTrip(req)
}
