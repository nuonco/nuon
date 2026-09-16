package skills

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

const indexURI = "skill:///index.md"

// Register adds the skill index/skill resources and the list_skills/load_skill
// tools to server. Resources are the primary surface for clients that browse
// MCP resources; the tools are a fallback for clients (like the CLI proxy)
// that only exercise tools, and their descriptions keep the catalog visible
// even to a model that never calls them.
func Register(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:         indexURI,
		Name:        "skill-index",
		Description: "Index of available agent skills, one line per skill. Load an individual skill via skill:///{domain}/{name}.md before acting on it.",
		MIMEType:    "text/markdown",
	}, readIndex)

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "skill:///{domain}/{name}.md",
		Name:        "skill",
		Description: "A single focused agent skill, scoped to a domain.",
		MIMEType:    "text/markdown",
	}, readSkill)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_skills",
		"List skills",
		"List available agent skills (one line each: domain/name and when to use it). Call this before writing or updating an app config, README, or runbook to check for a relevant skill, then call load_skill for the full guide.",
	), mcpListSkills)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"load_skill",
		"Load skill",
		"Load the full body of one skill by domain and name (from list_skills).",
	), mcpLoadSkill)
}

func readIndex(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      indexURI,
			MIMEType: "text/markdown",
			Text:     Index(),
		}},
	}, nil
}

func readSkill(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	domain, name, ok := parseSkillURI(req.Params.URI)
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	skill, ok := Find(domain, name)
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      skill.URI(),
			MIMEType: "text/markdown",
			Text:     skill.Body,
		}},
	}, nil
}

// parseSkillURI extracts domain/name from skill:///{domain}/{name}.md.
func parseSkillURI(uri string) (domain, name string, ok bool) {
	path := strings.TrimPrefix(uri, "skill:///")
	if path == uri {
		return "", "", false
	}
	path = strings.TrimSuffix(path, ".md")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

type listSkillsInput struct{}

type skillSummary struct {
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func mcpListSkills(_ context.Context, _ *mcp.CallToolRequest, _ listSkillsInput) (*mcp.CallToolResult, any, error) {
	summaries := make([]skillSummary, 0, len(All()))
	for _, s := range All() {
		summaries = append(summaries, skillSummary{Domain: s.Domain, Name: s.Name, Description: s.Description})
	}
	result, _, err := apiPkg.MCPJSONResult(summaries)
	return result, nil, err
}

type loadSkillInput struct {
	Domain string `json:"domain" jsonschema:"the skill's domain, e.g. apps"`
	Name   string `json:"name" jsonschema:"the skill's name, e.g. update-app-config-readme"`
}

func mcpLoadSkill(_ context.Context, _ *mcp.CallToolRequest, in loadSkillInput) (*mcp.CallToolResult, any, error) {
	skill, ok := Find(in.Domain, in.Name)
	if !ok {
		return nil, nil, fmt.Errorf("no skill %q/%q", in.Domain, in.Name)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: skill.Body}},
	}, nil, nil
}
