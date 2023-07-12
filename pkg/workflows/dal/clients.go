package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

const (
	defaultSessionTimeout  time.Duration = time.Second * 3600
	defaultSessionNameTmpl string        = "workflows-dal-%s-%s"
)

func (r *repo) getAppsCredentials(ctx context.Context) *credentials.Config {
	if r.Auth != nil {
		return r.Auth
	}

	return &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     fmt.Sprintf(r.Settings.AppsBucketIAMRoleTemplate, r.OrgId),
			SessionName: fmt.Sprintf(defaultSessionNameTmpl, "apps", r.OrgId),
		},
	}
}

func (r *repo) deploymentsCredentials(ctx context.Context) *credentials.Config {
	if r.Auth != nil {
		return r.Auth
	}

	return &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     fmt.Sprintf(r.Settings.DeploymentsBucketIAMRoleTemplate, r.OrgId),
			SessionName: fmt.Sprintf(defaultSessionNameTmpl, "deployments", r.OrgId),
		},
	}
}

func (r *repo) installsCredentials(ctx context.Context) *credentials.Config {
	if r.Auth != nil {
		return r.Auth
	}

	return &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     fmt.Sprintf(r.Settings.InstallsBucketIAMRoleTemplate, r.OrgId),
			SessionName: fmt.Sprintf(defaultSessionNameTmpl, "installs", r.OrgId),
		},
	}
}

func (r *repo) orgsCredentials(ctx context.Context) *credentials.Config {
	if r.Auth != nil {
		return r.Auth
	}

	return &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     fmt.Sprintf(r.Settings.OrgsBucketIAMRoleTemplate, r.OrgId),
			SessionName: fmt.Sprintf(defaultSessionNameTmpl, "orgs", r.OrgId),
		},
	}
}
