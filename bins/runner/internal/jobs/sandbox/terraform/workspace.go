package terraform

import (
	"fmt"

	"github.com/pkg/errors"

	dirarchive "github.com/nuonco/nuon/pkg/terraform/archive/dir"
	httpbackend "github.com/nuonco/nuon/pkg/terraform/backend/http"
	remotebinary "github.com/nuonco/nuon/pkg/terraform/binary/remote"
	"github.com/nuonco/nuon/pkg/terraform/hooks"
	"github.com/nuonco/nuon/pkg/terraform/hooks/noop"
	"github.com/nuonco/nuon/pkg/terraform/hooks/shell"
	authvars "github.com/nuonco/nuon/pkg/terraform/variables/auth"
	staticvars "github.com/nuonco/nuon/pkg/terraform/variables/static"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
)

// getWorkspace returns a valid workspace for working with this plugin
func (h *handler) getWorkspace() (workspace.Workspace, error) {
	plan := h.state.plan
	sandboxCfg := h.state.sandboxCfg

	archDir := h.state.workspace.Source().AbsPath()
	if plan.LocalArchive != nil {
		archDir = plan.LocalArchive.Path
	}

	arch, err := dirarchive.New(h.v,
		dirarchive.WithPath(archDir),
		dirarchive.WithIgnoreDotTerraformDir(),
		dirarchive.WithIgnoreTerraformStateFile(),
		dirarchive.WithAddBackendFile("http"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(h.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: h.cfg.RunnerAPIURL,
		WorkspaceID: h.state.plan.TerraformBackend.WorkspaceID,
		Token:       h.cfg.RunnerAPIToken,
		JobID:       h.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := remotebinary.New(h.v,
		remotebinary.WithVersion(sandboxCfg.TerraformVersion))
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(h.v,
		staticvars.WithFileVars(plan.Vars),
		staticvars.WithEnvVars(plan.EnvVars),
		staticvars.WithFiles(plan.VarsFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(h.v,
		authvars.WithAWSAuth(plan.AWSAuth),
		authvars.WithAzureAuth(plan.AzureAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	var hooks hooks.Hooks
	if plan.Hooks == nil {
		hooks = noop.New()
	} else {
		hooks, err = shell.New(h.v,
			shell.WithRunAuth(&plan.Hooks.RunAuth),
			shell.WithEnvVars(plan.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
		workspace.WithControlCache(true),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}

// getWorkspace returns a valid workspace for working with this plugin when we have a plan
// NOTE: these are kept distinct in case they continue to evolve separately and to make it easier to reason about
func (h *handler) getWorkspaceWithPlan(planBytes []byte) (workspace.Workspace, error) {
	plan := h.state.plan
	sandboxCfg := h.state.sandboxCfg

	archDir := h.state.workspace.Source().AbsPath()
	if plan.LocalArchive != nil {
		archDir = plan.LocalArchive.Path
	}

	arch, err := dirarchive.New(h.v,
		dirarchive.WithPath(archDir),
		dirarchive.WithIgnoreDotTerraformDir(),
		dirarchive.WithIgnoreTerraformStateFile(),
		dirarchive.WithAddBackendFile("http"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(h.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: h.cfg.RunnerAPIURL,
		WorkspaceID: h.state.plan.TerraformBackend.WorkspaceID,
		Token:       h.cfg.RunnerAPIToken,
		JobID:       h.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := remotebinary.New(h.v,
		remotebinary.WithVersion(sandboxCfg.TerraformVersion))
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(h.v,
		staticvars.WithFileVars(plan.Vars),
		staticvars.WithEnvVars(plan.EnvVars),
		staticvars.WithFiles(plan.VarsFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(h.v,
		authvars.WithAWSAuth(plan.AWSAuth),
		authvars.WithAzureAuth(plan.AzureAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	var hooks hooks.Hooks
	if plan.Hooks == nil {
		hooks = noop.New()
	} else {
		hooks, err = shell.New(h.v,
			shell.WithRunAuth(&plan.Hooks.RunAuth),
			shell.WithEnvVars(plan.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
		workspace.WithControlCache(true),
		workspace.WithPlanBytes(planBytes),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
