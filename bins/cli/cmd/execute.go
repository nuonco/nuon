package cmd

import (
	"context"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/powertoolsdev/mono/bins/cli/internal/builds"
	"github.com/powertoolsdev/mono/bins/cli/internal/components"
	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

// Execute is essentially the init method of the CLI. It initializes all the components and composes them together.
// TODO(ja): one could argue this and config.go should be the only files in cmd, and the commands, now that they're
// decoupled from this init code, could go in their own internal package. That's not the convention cobra promotes,
// but it could make sense since we're using constructors for all the commands.
func Execute() {
	// Construct a validator for the API client and the UI logger.
	vld := validator.New()

	// Construct an API client for the services to use.
	api, err := newAPI(vld)
	if err != nil {
		log.Fatal(err)
	}

	// Construct the domain services, and pass them to the CLI commands that use them.
	cmds := []*cobra.Command{
		newAppsCmd(bindConfig, apps.New(api)),
		newComponentsCmd(bindConfig, components.New(api)),
		newInstallsCmd(bindConfig, installs.New(api)),
		newOrgsCmd(bindConfig, orgs.New(api)),
		newVersionCmd(bindConfig, version.New()),
		newBuildsCmd(bindConfig, builds.New(api)),
	}

	// Create a context to pass down the code path.
	ctx := context.Background()
	uiLog, err := ui.New(vld)
	if err != nil {
		log.Fatalf("unable to initialize ui: %s", err)
	}
	ctx = ui.WithContext(ctx, uiLog)

	// Construct and init the root command.
	rootCmd := newRootCmd(bindConfig, cmds...)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(2)
	}
}
