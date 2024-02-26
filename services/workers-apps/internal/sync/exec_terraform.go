package sync

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/terraform/archive/json"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/backend/local"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

const (
	defaultFileName        string = "module.tf.json"
	localStateFileTemplate string = "/tmp/%s-terraform.tfstate"
)

type ExecTerraformRequest struct {
	TerraformJSON    string `validate:"required"`
	TerraformVersion string `validate:"required"`

	// values to run terraform with
	APIURL   string `validate:"required"`
	OrgID    string `validate:"required"`
	AppID    string `validate:"required"`
	APIToken string `validate:"required"`

	BackendBucket     string `validate:"required"`
	BackendKey        string `validate:"required"`
	BackendRegion     string `validate:"required"`
	BackendIAMRoleARN string `validate:"required"`
}

func (a *Activities) ExecTerraform(ctx context.Context, req *ExecTerraformRequest) error {
	wkspace, err := a.getWorkspace(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to get workspace: %w", err)
	}

	runLog := hclog.New(&hclog.LoggerOptions{
		Name:   "terraform",
		Output: os.Stdout,
	})

	runUI := terminal.NonInteractiveUI(ctx)
	tfRun, err := run.New(a.v,
		run.WithWorkspace(wkspace),
		run.WithUI(runUI),
		run.WithLogger(runLog),
		run.WithOutputSettings(&run.OutputSettings{
			Ignore: true,
		}),
	)
	if err != nil {
		return fmt.Errorf("unable to create run: %w", err)
	}

	if err := tfRun.Apply(ctx); err != nil {
		return fmt.Errorf("unable to apply seed terraform: %w", err)
	}

	return nil
}

func (a *Activities) getWorkspace(ctx context.Context, req *ExecTerraformRequest) (workspace.Workspace, error) {
	arch, err := json.New(a.v,
		json.WithFileName(defaultFileName),
		json.WithJSON([]byte(req.TerraformJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create archive: %w", err)
	}

	bin, err := remotebinary.New(a.v,
		remotebinary.WithVersion(req.TerraformVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(a.v, staticvars.WithFileVars(map[string]interface{}{
		"app_id": req.AppID,
	}),
		staticvars.WithEnvVars(map[string]string{
			"NUON_ORG_ID":  req.OrgID,
			"NUON_API_URL": req.APIURL,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create vars: %w", err)
	}

	var back backend.Backend
	back, err = s3backend.New(a.v,
		s3backend.WithCredentials(&credentials.Config{
			AssumeRole: &credentials.AssumeRoleConfig{
				SessionName: "workers-app-sync",
				RoleARN:     req.BackendIAMRoleARN,
			},
		}),
		s3backend.WithBucketConfig(&s3backend.BucketConfig{
			Name:   req.BackendBucket,
			Key:    req.BackendKey,
			Region: req.BackendRegion,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	if a.cfg.Env == config.Development {
		stateFP := fmt.Sprintf(localStateFileTemplate, req.AppID)
		back, err = local.New(a.v, local.WithFilepath(stateFP))
		if err != nil {
			return nil, fmt.Errorf("unable to create local backend: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create local backend: %w", err)
	}

	hooks := noop.New()

	// create workspace
	wkspace, err := workspace.New(a.v,
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
