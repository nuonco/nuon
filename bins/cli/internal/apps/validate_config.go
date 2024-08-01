package apps

import (
	"context"
	"fmt"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/schema"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	localStateFileTemplate string = "/tmp/%s-terraform.tfstate"
)

func (s *Service) loadConfig(ctx context.Context, file string) (*config.AppConfig, error) {
	cfg, err := parse.Parse(parse.ParseConfig{
		Context:     config.ConfigContextSource,
		Filename:    file,
		BackendType: config.BackendTypeLocal,
		Template:    true,
		V:           validator.New(),
	})
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s *Service) Validate(ctx context.Context, all bool, file string, asJSON bool) {
	view := ui.NewGetView()

	var (
		cfgFiles []parse.File
		err      error
	)

	if all {
		view.Print("searching for config files in current directory")
		cfgFiles, err = parse.FindConfigFiles(".")
		if err != nil {
			view.Error(err)
			return
		}
	}
	if file != "" {
		appName, err := parse.AppNameFromFilename(file)
		if err != nil {
			view.Error(err)
			return
		}

		view.Print(fmt.Sprintf("found %s", file))
		cfgFiles = []parse.File{
			{
				Path:    file,
				AppName: appName,
			},
		}
	}

	if len(cfgFiles) < 1 {
		view.Error(&ui.CLIUserError{
			Msg: fmt.Sprintf("must set --all or --file, and make sure at least one \"nuon.<app-name>.toml\" file exists"),
		})
		return
	}

	for _, cfgFile := range cfgFiles {
		view.Print(fmt.Sprintf("validating file \"%s\"", cfgFile.Path))
		if err := s.validate(ctx, cfgFile, asJSON); err != nil {
			view.Error(err)
			break
		}
	}
}

func (s *Service) validate(ctx context.Context, file parse.File, asJSON bool) error {
	view := ui.NewListView()

	cfg, err := s.loadConfig(ctx, file.Path)
	if err != nil {
		ui.PrintError(err)
		return err
	}

	if err := cfg.Validate(s.v); err != nil {
		ui.PrintError(err)
		return err
	}

	schmaErrs, err := schema.Validate(cfg)
	if err != nil {
		ui.PrintError(err)
		return err
	}

	if len(schmaErrs) < 1 {
		ui.PrintSuccess("successfully validated " + file.Path)
		return nil
	}

	view.Print(fmt.Sprintf("%d total errors", len(schmaErrs)))
	for _, schemaErr := range schmaErrs {
		view.Print(schemaErr.String())
	}

	err = s.validateDuplicateComponentNames(cfg)
	if err != nil {
		ui.PrintError(err)
	}

	return nil
}

func (s *Service) validateDuplicateComponentNames( cfg *config.AppConfig) error {
	componentNames := make(map[string]bool)
	for _, v := range cfg.Components {
		if _, ok := componentNames[v.Name]; ok {
			return errs.UserFacingError(
				errors.New("duplicate component name"),
				fmt.Sprintf("Validation error: duplicate component name %q", v.Name),
			)
		}
		componentNames[v.Name] = true
	}
	return nil
}
