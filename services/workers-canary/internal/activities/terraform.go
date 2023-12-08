package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/terraform/archive/dir"
	"github.com/powertoolsdev/mono/pkg/terraform/backend/local"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	"github.com/powertoolsdev/mono/pkg/terraform/outputs"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

const (
	defaultTerraformVersion string = "v1.5.3"
)

type RunType string

const (
	RunTypeDestroy RunType = "destroy"
	RunTypeApply   RunType = "apply"
)

type TerraformRunOutputs struct {
	OrgID string `mapstructure:"org_id" json:"org_id"`

	AppID string                 `mapstructure:"app_id" json:"app_id"`
	App   map[string]interface{} `mapstructure:"app" json:"app"`

	ComponentIDs []string               `mapstructure:"component_ids" json:"component_ids"`
	Components   map[string]interface{} `mapstructure:"components" json:"components"`

	InstallIDs []string               `mapstructure:"install_ids" json:"install_ids"`
	Installs   map[string]interface{} `mapstructure:"installs" json:"installs"`
}

type RunTerraformRequest struct {
	RunType  RunType
	CanaryID string
	OrgID    string
}

type RunTerraformResponse struct {
	Outputs *TerraformRunOutputs
}

func (a *Activities) getWorkspace(moduleDir, canaryID, orgID string) (workspace.Workspace, error) {
	arch, err := dir.New(a.v,
		dir.WithPath(moduleDir),
		dir.WithIgnoreDotTerraformDir(),
		dir.WithIgnoreTerraformStateFile(),
		dir.WithIgnoreTerraformLockFile(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create archive: %w", err)
	}

	bin, err := remotebinary.New(a.v,
		remotebinary.WithVersion(defaultTerraformVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(a.v, staticvars.WithFileVars(map[string]interface{}{
		"install_role_arn": a.cfg.InstallIamRoleArn,
	}),
		staticvars.WithEnvVars(map[string]string{
			"NUON_ORG_ID":  orgID,
			"NUON_API_URL": a.cfg.APIURL,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create vars: %w", err)
	}

	stateFp := filepath.Join(a.cfg.TerraformStateBaseDir, fmt.Sprintf("%s-terraform.tfstate", canaryID))
	back, err := local.New(a.v, local.WithFilepath(stateFp))
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

func (a *Activities) runTerraform(ctx context.Context, moduleDir, canaryID, orgID string, runTyp RunType) (map[string]interface{}, error) {
	wkspace, err := a.getWorkspace(moduleDir, canaryID, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace: %w", err)
	}

	runLog := hclog.New(&hclog.LoggerOptions{
		Name:   "terraform",
		Output: os.Stderr,
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
		return nil, fmt.Errorf("unable to create run: %w", err)
	}

	if runTyp == RunTypeDestroy {
		err = tfRun.Destroy(ctx)
	} else {
		err = tfRun.Apply(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to run: %w", err)
	}

	wkspaceOutputs, err := wkspace.Output(ctx, runLog)
	if err != nil {
		return nil, fmt.Errorf("unable to get outputs: %w", err)
	}
	out, err := outputs.TFOutputMetaToStructPB(wkspaceOutputs)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to standard outputs: %w", err)
	}

	return out.AsMap(), nil
}

func (a *Activities) RunTerraform(ctx context.Context, req *RunTerraformRequest) (*RunTerraformResponse, error) {
	moduleDir := a.cfg.TerraformModuleDir
	rawOutputs, err := a.runTerraform(ctx, moduleDir, req.CanaryID, req.OrgID, req.RunType)
	if err != nil {
		return nil, fmt.Errorf("unable to run terraform: %w", err)
	}

	var outputs TerraformRunOutputs
	if err := mapstructure.Decode(rawOutputs, &outputs); err != nil {
		return nil, fmt.Errorf("unable to parse outputs: %w", err)
	}
	outputs.OrgID = req.OrgID

	return &RunTerraformResponse{
		Outputs: &outputs,
	}, nil
}
