package apps

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/terraform/archive/json"
	"github.com/powertoolsdev/mono/pkg/terraform/backend/local"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

const (
	localStateFileTemplate string = "/tmp/%s-terraform.tfstate"
)

func (s *Service) loadConfig(ctx context.Context, file string) ([]byte, error) {
	byts, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tfJSON, err := parse.ToTerraformJSON(parse.ParseConfig{
		Bytes:       byts,
		BackendType: config.BackendTypeLocal,
		Template:    true,
		V:           validator.New(),
	})
	if err != nil {
		return nil, err
	}

	return tfJSON, nil
}

func (a *Service) getWorkspace(ctx context.Context, appID string, tfJSON []byte) (workspace.Workspace, error) {
	arch, err := json.New(a.v,
		json.WithFileName(config.DefaultModuleFileName),
		json.WithJSON(tfJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create archive: %w", err)
	}

	bin, err := remotebinary.New(a.v,
		remotebinary.WithVersion(config.DefaultTerraformVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(a.v, staticvars.WithFileVars(map[string]interface{}{
		"app_id": appID,
	}),
		staticvars.WithEnvVars(map[string]string{
			"NUON_ORG_ID":    a.cfg.OrgID,
			"NUON_API_URL":   a.cfg.APIURL,
			"NUON_API_TOKEN": a.cfg.APIToken,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create vars: %w", err)
	}

	stateFP := fmt.Sprintf(localStateFileTemplate, uuid.New())
	back, err := local.New(a.v, local.WithFilepath(stateFP))
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

func (a *Service) execTerraformValidate(ctx context.Context, appID string, tfJSON []byte) (*tfjson.ValidateOutput, error) {
	wkspace, err := a.getWorkspace(ctx, appID, tfJSON)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace: %w", err)
	}

	output := io.Discard
	if os.Getenv("NUON_DEBUG") != "" {
		output = os.Stdout
	}
	runLog := hclog.New(&hclog.LoggerOptions{
		Name:   "terraform",
		Output: output,
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

	if err := tfRun.Validate(ctx); err != nil {
		return nil, err
	}

	validateOut, err := wkspace.Validate(ctx, runLog)
	if err != nil {
		return nil, err
	}

	return validateOut, nil
}

func (s *Service) Validate(ctx context.Context, file string, asJSON bool) {
	view := ui.NewGetView()

	appName, err := parse.AppNameFromFilename(file)
	if err != nil {
		ui.PrintError(err)
		return
	}
	appID, err := lookup.AppID(ctx, s.api, appName)
	if err != nil {
		ui.PrintError(err)
		return
	}

	tfJSON, err := s.loadConfig(ctx, file)
	if err != nil {
		ui.PrintError(err)
		return
	}

	validateOutput, err := s.execTerraformValidate(ctx, appID, tfJSON)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if asJSON {
		ui.PrintJSON(validateOutput)
		return
	}

	if len(validateOutput.Diagnostics) < 1 {
		view.Print("ok")
		return
	}

	data := [][]string{
		{"resource", "summary", "error"},
	}
	for _, diag := range validateOutput.Diagnostics {
		data = append(data, []string{
			*diag.Snippet.Context,
			diag.Summary,
			diag.Detail,
		})
	}

	view.Render(data)
	view.Print(fmt.Sprintf("%d errors", len(validateOutput.Diagnostics)))
}
