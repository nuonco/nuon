package platform

import (
	"fmt"

	ociarchive "github.com/powertoolsdev/mono/pkg/terraform/archive/oci"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// GetWorkspace returns a valid workspace for working with this plugin
func (p *Platform) GetWorkspace() (workspace.Workspace, error) {
	arch, err := ociarchive.New(p.v)
	if err != nil {
		return nil, fmt.Errorf("unable to create archive: %w", err)
	}

	back, err := s3backend.New(p.v, s3backend.WithCredentials(&s3backend.Credentials{
		AWSAccessKeyID:     p.Cfg.Backend.Auth.AccessKeyID,
		AWSSecretAccessKey: p.Cfg.Backend.Auth.SecretAccessKey,
		AWSSessionToken:    p.Cfg.Backend.Auth.SessionToken,
	}), s3backend.WithBucketConfig(&s3backend.BucketConfig{
		Name:   p.Cfg.Backend.Bucket,
		Key:    p.Cfg.Backend.StateKey,
		Region: "us-west-2",
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	bin, err := remotebinary.New(p.v, remotebinary.WithVersion(p.Cfg.TerraformVersion))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	vars, err := staticvars.New(p.v)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	wkspace, err := workspace.New(p.v,
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
