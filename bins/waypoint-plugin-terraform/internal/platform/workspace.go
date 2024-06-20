package platform

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	dirarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/dir"
	ociarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/oci"
	s3archive "github.com/powertoolsdev/mono/pkg/terraform/archive/s3"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/shell"
	authvars "github.com/powertoolsdev/mono/pkg/terraform/variables/auth"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// GetWorkspace returns a valid workspace for working with this plugin
func (p *Platform) GetWorkspace() (workspace.Workspace, error) {
	var arch archive.Archive

	if p.Cfg.OCIArchive != nil {
		var err error
		arch, err = ociarchive.New(p.v, ociarchive.WithAuth(&ociarchive.Auth{
			Username: p.Cfg.OCIArchive.Username,
			Token:    p.Cfg.OCIArchive.AuthToken,
		}), ociarchive.WithImage(&ociarchive.Image{
			Registry: p.Cfg.OCIArchive.RegistryURL,
			Repo:     p.Cfg.OCIArchive.Repo,
			Tag:      p.Cfg.OCIArchive.Tag,
		}))
		if err != nil {
			return nil, fmt.Errorf("unable to create oci archive: %w", err)
		}
	} else if p.Cfg.S3Archive != nil {
		var err error
		arch, err = s3archive.New(p.v,
			s3archive.WithCredentials(&p.Cfg.S3Archive.Auth),
			s3archive.WithBucketKey(p.Cfg.S3Archive.BucketKey),
			s3archive.WithBucketName(p.Cfg.S3Archive.Bucket),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create archive: %w", err)
		}
	} else if p.Cfg.DirArchive != nil {
		var err error
		arch, err = dirarchive.New(p.v,
			dirarchive.WithPath(filepath.Join(p.Path, p.Cfg.DirArchive.Path)),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create local archive: %w", err)
		}
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

	cfgVars := generics.ToIntMap(p.Cfg.Variables)
	if p.Cfg.VariablesJSON != "" {
		var jsonVars map[string]interface{}
		if err := json.Unmarshal([]byte(p.Cfg.VariablesJSON), &cfgVars); err != nil {
			return nil, fmt.Errorf("unable to parse json vars: %w", err)
		}
		cfgVars = generics.MergeMap(cfgVars, jsonVars)
	}

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(cfgVars),
		staticvars.WithEnvVars(p.Cfg.EnvVars))
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(p.v,
		authvars.WithAuth(&p.Cfg.RunAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	var hooks hooks.Hooks
	if p.Cfg.Hooks == nil {
		hooks = noop.New()
	} else {
		hooks, err = shell.New(p.v,
			shell.WithRunAuth(&p.Cfg.Hooks.RunAuth),
			shell.WithEnvVars(p.Cfg.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

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
