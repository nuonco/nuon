package terraform

import (
	"fmt"

	"github.com/pkg/errors"

	dirarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/dir"
	httpbackend "github.com/powertoolsdev/mono/pkg/terraform/backend/http"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	authvars "github.com/powertoolsdev/mono/pkg/terraform/variables/auth"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// GetWorkspace returns a valid workspace for working with this plugin
func (p *handler) GetWorkspace() (workspace.Workspace, error) {
	arch, err := dirarchive.New(p.v,
		dirarchive.WithPath(p.state.arch.BasePath()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(p.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: p.cfg.RunnerAPIURL,
		WorkspaceID: p.state.plan.TerraformDeployPlan.TerraformBackend.WorkspaceID,
		Token:       p.cfg.RunnerAPIToken,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := remotebinary.New(p.v,
		remotebinary.WithVersion(p.state.terraformCfg.Version))
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(p.state.plan.TerraformDeployPlan.Vars),
		staticvars.WithEnvVars(p.state.plan.TerraformDeployPlan.EnvVars))
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(p.v,
		authvars.WithAWSAuth(p.state.plan.TerraformDeployPlan.AWSAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	hooks := noop.New()

	wkspace, err := workspace.New(p.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
