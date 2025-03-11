package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	dirarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/dir"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/shell"
	authvars "github.com/powertoolsdev/mono/pkg/terraform/variables/auth"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// getWorkspace returns a valid workspace for working with this plugin
func (h *handler) getWorkspace() (workspace.Workspace, error) {
	cfg := h.state.cfg

	archDir := h.state.workspace.Source().AbsPath()
	if cfg.DirArchive != nil {
		archDir = cfg.DirArchive.Path
	}

	arch, err := dirarchive.New(h.v,
		dirarchive.WithPath(archDir),
		dirarchive.WithIgnoreDotTerraformDir(),
		dirarchive.WithIgnoreTerraformStateFile(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := s3backend.New(h.v, s3backend.WithCredentials(&cfg.Backend.Auth), s3backend.WithBucketConfig(&s3backend.BucketConfig{
		Name: cfg.Backend.Bucket, Key: cfg.Backend.StateKey,
		Region: cfg.Backend.Region,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	bin, err := remotebinary.New(h.v,
		remotebinary.WithVersion(cfg.TerraformVersion))
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	cfgVars := generics.ToIntMap(cfg.Variables)
	if cfg.VariablesJSON != "" {
		var jsonVars map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.VariablesJSON), &cfgVars); err != nil {
			return nil, fmt.Errorf("unable to parse json vars: %w", err)
		}
		cfgVars = generics.MergeMap(cfgVars, jsonVars)
	}

	vars, err := staticvars.New(h.v,
		staticvars.WithFileVars(cfgVars),
		staticvars.WithEnvVars(cfg.EnvVars))
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(h.v,
		authvars.WithAWSAuth(&cfg.RunAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	var hooks hooks.Hooks
	if cfg.Hooks == nil {
		hooks = noop.New()
	} else {
		hooks, err = shell.New(h.v,
			shell.WithRunAuth(&cfg.Hooks.RunAuth),
			shell.WithEnvVars(cfg.Hooks.EnvVars),
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
