package cmd

import (
	"context"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/powertoolsdev/mono/bins/cli/internal/components"
	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

// Execute is essentially the init method of the CLI. It initializes all the components and composes them together.
func Execute() {
	// Create the root command that all other commands will be nested under.
	// If there are any flags or other settings that we want to be "global",
	// they should be configured on this command.
	rootCmd := &cobra.Command{
		Use:          "nuonctl",
		SilenceUsage: true,
		// TODO(ja): PersistentPreRunE is only inherited by immediate child commands,
		// so we still have to set this on each subcommand, so that it's children
		// will inherit it. There are a couple ways we can refactor things to avoid
		// this, but for now I'm just copy/pasting.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	// Construct a validator for the API client and the UI logger.
	vld := validator.New()

	// Construct an API client for the services to use.
	api, err := newAPI(vld)
	if err != nil {
		log.Fatal(err)
	}

	// Construct the domain services, and pass them to the CLI commands that use them.
	appsCmd := registerApps(apps.New(api))
	rootCmd.AddCommand(&appsCmd)

	componentsCmd := registerComponents(components.New(api))
	rootCmd.AddCommand(&componentsCmd)

	installsCmd := registerInstalls(installs.New(api))
	rootCmd.AddCommand(&installsCmd)

	orgsCmd := registerOrgs(orgs.New(api))
	rootCmd.AddCommand(&orgsCmd)

	versionCmd := registerVersion(version.New())
	rootCmd.AddCommand(&versionCmd)

	// Create a context to pass down the code path.
	ctx := context.Background()
	uiLog, err := ui.New(vld)
	if err != nil {
		log.Fatalf("unable to initialize ui: %s", err)
	}
	ctx = ui.WithContext(ctx, uiLog)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(2)
	}
}
