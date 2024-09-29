package terraform

import (
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/generics"
	dirarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/dir"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	authvars "github.com/powertoolsdev/mono/pkg/terraform/variables/auth"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// getWorkspace returns a valid workspace for working with this plugin
func (p *handler) getWorkspace() (workspace.Workspace, error) {
	cfg := p.state.cfg

	bundleDir := filepath.Join(p.cfg.BundleDir, cfg.BundleName)
	arch, err := dirarchive.New(p.v,
		dirarchive.WithPath(bundleDir),
		dirarchive.WithIgnoreDotTerraformDir(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := s3backend.New(p.v,
		s3backend.WithCredentials(&cfg.Backend.Auth),
		s3backend.WithBucketConfig(&s3backend.BucketConfig{
			Name:   cfg.Backend.Bucket,
			Key:    cfg.Backend.StateKey,
			Region: cfg.Backend.Region,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	bin, err := remotebinary.New(p.v,
		remotebinary.WithVersion(cfg.TerraformVersion))
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	cfgVars := generics.ToIntMap(cfg.Variables)

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(cfgVars),
		staticvars.WithEnvVars(cfg.EnvVars))
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	// set the authentication variables from the config
	authVars, err := authvars.New(p.v,
		authvars.WithAWSAuth(cfg.AWSAuth),
		authvars.WithAzureAuth(cfg.AzureAuth),
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
