package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/paginate"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// jsonResult marshals v into text content. Models are trimmed before this
// point: full API objects are far too large for an LLM context.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(j)}},
	}, nil, nil
}

type appSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toAppSummary(a *models.AppApp) appSummary {
	return appSummary{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type installSummary struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	AppID           string `json:"app_id"`
	SandboxStatus   string `json:"sandbox_status,omitempty"`
	RunnerStatus    string `json:"runner_status,omitempty"`
	ComponentStatus string `json:"component_status,omitempty"`
	CloudPlatform   string `json:"cloud_platform,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toInstallSummary(i *models.AppInstall) installSummary {
	return installSummary{
		ID:              i.ID,
		Name:            i.Name,
		AppID:           i.AppID,
		SandboxStatus:   i.SandboxStatus,
		RunnerStatus:    i.RunnerStatus,
		ComponentStatus: i.CompositeComponentStatus,
		CloudPlatform:   i.CloudPlatform,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}
}

type componentSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AppID          string `json:"app_id"`
	ConfigVersions int64  `json:"config_versions"`
	CreatedAt      string `json:"created_at"`
}

func toComponentSummary(c *models.AppComponent) componentSummary {
	return componentSummary{
		ID:             c.ID,
		Name:           c.Name,
		AppID:          c.AppID,
		ConfigVersions: c.ConfigVersions,
		CreatedAt:      c.CreatedAt,
	}
}

type installComponentSummary struct {
	ComponentID    string `json:"component_id"`
	Name           string `json:"name"`
	DeployStatus   string `json:"deploy_status,omitempty"`
	LatestDeployID string `json:"latest_deploy_id,omitempty"`
}

func toInstallComponentSummary(c *models.AppInstallComponent) installComponentSummary {
	out := installComponentSummary{}
	if c.Component != nil {
		out.ComponentID = c.Component.ID
		out.Name = c.Component.Name
	}
	if len(c.InstallDeploys) > 0 {
		out.DeployStatus = c.InstallDeploys[0].Status
		out.LatestDeployID = c.InstallDeploys[0].ID
	}
	return out
}

type emptyInput struct{}

type appInput struct {
	App string `json:"app,omitempty" jsonschema:"app name or ID; defaults to the currently selected app"`
}

type optionalAppInput struct {
	App string `json:"app,omitempty" jsonschema:"app name or ID; defaults to the currently selected app"`
	All bool   `json:"all,omitempty" jsonschema:"set true to list across the whole org, ignoring the selected app"`
}

type installInput struct {
	Install string `json:"install,omitempty" jsonschema:"install name or ID; defaults to the currently selected install"`
}

// resolveApp falls back to the CLI's selected app (nuon apps select) when no
// app is given, matching CLI command behavior.
func (s *Service) resolveApp(ctx context.Context, app string) (string, error) {
	if app == "" && s.cfg != nil {
		app = s.cfg.AppID
	}
	return lookup.AppID(ctx, s.api, app)
}

func (s *Service) resolveInstall(ctx context.Context, install string) (string, error) {
	if install == "" && s.cfg != nil {
		install = s.cfg.InstallID
	}
	return lookup.InstallID(ctx, s.api, install)
}

func (s *Service) registerReadTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Get the current authenticated user, org, and the selected app/install context. Other tools default to this context when their app/install argument is omitted.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		user, err := s.api.GetCurrentUser(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := s.api.GetOrg(ctx)
		if err != nil {
			return nil, nil, err
		}
		out := map[string]any{
			"user": map[string]string{"id": user.ID, "email": user.Email},
			"org":  map[string]string{"id": org.ID, "name": org.Name},
		}
		if s.cfg != nil && s.cfg.AppID != "" {
			selected := map[string]string{"id": s.cfg.AppID}
			if app, err := s.api.GetApp(ctx, s.cfg.AppID); err == nil {
				selected["id"] = app.ID
				selected["name"] = app.Name
			}
			out["selected_app"] = selected
		}
		if s.cfg != nil && s.cfg.InstallID != "" {
			selected := map[string]string{"id": s.cfg.InstallID}
			if install, err := s.api.GetInstall(ctx, s.cfg.InstallID); err == nil {
				selected["id"] = install.ID
				selected["name"] = install.Name
			}
			out["selected_install"] = selected
		}
		return jsonResult(out)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_apps",
		Description: "List all apps in the current org.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		apps, err := paginate.All(func(off, lim int) ([]*models.AppApp, bool, error) {
			return s.api.GetApps(ctx, &models.GetPaginatedQuery{Offset: off, Limit: lim})
		})
		if err != nil {
			return nil, nil, err
		}
		out := make([]appSummary, 0, len(apps))
		for _, a := range apps {
			out = append(out, toAppSummary(a))
		}
		return jsonResult(out)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_app",
		Description: "Get an app by name or ID; defaults to the currently selected app.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in appInput) (*mcp.CallToolResult, any, error) {
		appID, err := s.resolveApp(ctx, in.App)
		if err != nil {
			return nil, nil, err
		}
		app, err := s.api.GetApp(ctx, appID)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(toAppSummary(app))
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_installs",
		Description: "List installs for the selected app (or the app given); pass all=true for every install in the org.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in optionalAppInput) (*mcp.CallToolResult, any, error) {
		fetch := func(off, lim int) ([]*models.AppInstall, bool, error) {
			return s.api.GetAllInstalls(ctx, &models.GetPaginatedQuery{Offset: off, Limit: lim})
		}
		if !in.All {
			if appID, err := s.resolveApp(ctx, in.App); err == nil {
				fetch = func(off, lim int) ([]*models.AppInstall, bool, error) {
					return s.api.GetAppInstalls(ctx, appID, &models.GetPaginatedQuery{Offset: off, Limit: lim})
				}
			} else if in.App != "" {
				return nil, nil, err
			}
		}
		installs, err := paginate.All(fetch)
		if err != nil {
			return nil, nil, err
		}
		out := make([]installSummary, 0, len(installs))
		for _, i := range installs {
			out = append(out, toInstallSummary(i))
		}
		return jsonResult(out)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_install",
		Description: "Get an install by name or ID (defaults to the currently selected install), including sandbox/runner/component status.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in installInput) (*mcp.CallToolResult, any, error) {
		installID, err := s.resolveInstall(ctx, in.Install)
		if err != nil {
			return nil, nil, err
		}
		install, err := s.api.GetInstall(ctx, installID)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(toInstallSummary(install))
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_install_components",
		Description: "List the components on an install (defaults to the currently selected install) with their latest deploy status.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in installInput) (*mcp.CallToolResult, any, error) {
		installID, err := s.resolveInstall(ctx, in.Install)
		if err != nil {
			return nil, nil, err
		}
		comps, err := paginate.All(func(off, lim int) ([]*models.AppInstallComponent, bool, error) {
			return s.api.GetInstallComponents(ctx, installID, &models.GetPaginatedQuery{Offset: off, Limit: lim})
		})
		if err != nil {
			return nil, nil, err
		}
		out := make([]installComponentSummary, 0, len(comps))
		for _, c := range comps {
			out = append(out, toInstallComponentSummary(c))
		}
		return jsonResult(out)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_components",
		Description: "List components for the selected app (or the app given); pass all=true for every component in the org.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in optionalAppInput) (*mcp.CallToolResult, any, error) {
		fetch := func(off, lim int) ([]*models.AppComponent, bool, error) {
			return s.api.GetAllComponents(ctx, &models.GetPaginatedQuery{Offset: off, Limit: lim})
		}
		if !in.All {
			if appID, err := s.resolveApp(ctx, in.App); err == nil {
				fetch = func(off, lim int) ([]*models.AppComponent, bool, error) {
					return s.api.GetAppComponents(ctx, appID, &models.GetPaginatedQuery{Offset: off, Limit: lim})
				}
			} else if in.App != "" {
				return nil, nil, err
			}
		}
		comps, err := paginate.All(fetch)
		if err != nil {
			return nil, nil, err
		}
		out := make([]componentSummary, 0, len(comps))
		for _, c := range comps {
			out = append(out, toComponentSummary(c))
		}
		return jsonResult(out)
	})
}
