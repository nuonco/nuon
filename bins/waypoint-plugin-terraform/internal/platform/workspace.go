package platform

import (
	"fmt"

	ociarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/oci"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// GetWorkspace returns a valid workspace for working with this plugin
func (p *Platform) GetWorkspace() (workspace.Workspace, error) {
	arch, err := ociarchive.New(p.v, ociarchive.WithAuth(&ociarchive.Auth{
		Username: p.Cfg.Archive.Username,
		Token:    p.Cfg.Archive.AuthToken,
	}), ociarchive.WithImage(&ociarchive.Image{
		Registry: p.Cfg.Archive.RegistryURL,
		Repo:     p.Cfg.Archive.Repo,
		Tag:      p.Cfg.Archive.Tag,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to create archive: %w", err)
	}

	back, err := s3backend.New(p.v, s3backend.WithCredentials(&p.Cfg.Backend.Auth),
		s3backend.WithBucketConfig(&s3backend.BucketConfig{
			Name:   p.Cfg.Backend.Bucket,
			Key:    p.Cfg.Backend.StateKey,
			Region: p.Cfg.Backend.Region,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	bin, err := remotebinary.New(p.v,
		remotebinary.WithVersion(p.Cfg.TerraformVersion))
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	cfgVars := make(map[string]interface{})
	for k, v := range p.Cfg.Variables {
		cfgVars[k] = v
	}

	vars, err := staticvars.New(p.v, staticvars.WithFileVars(cfgVars))
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	hooks := noop.New()
	wkspace, err := workspace.New(p.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
